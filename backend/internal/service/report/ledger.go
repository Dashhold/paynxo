// Package report implements the Report_Service (Req 10, 11): tenant-scoped
// ledger aggregation and (in later tasks) the six report types and CSV/PDF
// exports. It reuses the Commission_Engine (internal/service/commission) over
// records read through the tenant + owner scope (repo.ScopeTenant) so every
// computed figure is restricted to the requesting principal's tenant and
// permitted scope (Req 10.5).
//
// This file implements task 10.1: ledger aggregation. The report-generation
// and export functions (tasks 10.2 / 10.3) extend this package without
// changing the ledger logic here.
//
// Parity with calc.js. The three ledgers mirror calc.js companyLedger,
// affiliateLedger, and merchantLedger field-for-field:
//
//   - Company:   receivable = Σ companyNetIncome over the company's txns;
//                paid       = Σ settlements to the company;
//                balance    = receivable - paid.
//   - Affiliate: earned  = Σ affiliateCommission over the affiliate's
//                          merchants' txns — the affiliate's gross cut LESS the
//                          merchant share carved out of it; paid = Σ affiliate
//                          payments; balance = earned - paid.
//   - Merchant:  earned  = Σ merchantCommission over the merchant's txns. Every
//                          merchant earns its own commission: a direct merchant
//                          is paid it straight out of the admin's commission,
//                          and an affiliate-assigned merchant has it carved out
//                          of its affiliate's cut (Req 10.4). paid = Σ merchant
//                          payments; balance = earned - paid.
package report

import (
	"errors"

	"gorm.io/gorm"

	"pgcs/backend/internal/model"
	"pgcs/backend/internal/repo"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/commission"
)

// Ledger is the aggregated balance returned for a company, affiliate, or
// merchant. A company ledger populates Receivable; an affiliate or merchant
// ledger populates Earned. Paid and Balance are populated for all three.
type Ledger struct {
	Receivable float64 `json:"receivable"`
	Earned     float64 `json:"earned"`
	Paid       float64 `json:"paid"`
	Balance    float64 `json:"balance"`
}

// Service computes tenant-scoped ledgers from the database. It holds only the
// GORM handle; the per-request tenant/owner scope is supplied through the
// service.Principal passed to each method.
type Service struct {
	db *gorm.DB
}

// NewService constructs a report Service bound to a database handle.
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// scoped returns a session that applies the principal's tenant + owner scope
// (repo.ScopeTenant) to the aggregation query. It is used for the tables that
// carry the principal's owner column (transactions/settlements have company_id
// and merchant_id; merchants/affiliate_payments have affiliate_id), so portal
// principals can only aggregate over the records they own within their tenant
// (Req 10.5, 7.5).
func (s *Service) scoped(p service.Principal) *gorm.DB {
	return s.db.Scopes(repo.ScopeTenant(p))
}

// CompanyLedger computes the ledger for the company with the given id within
// the principal's scope (Req 10.1). Receivable is the sum of company net
// income across the company's transactions (computed by the Commission_Engine
// from the company's gateway assignments); paid is the sum of settlements to
// the company; balance is receivable minus paid.
func (s *Service) CompanyLedger(p service.Principal, companyID string) (Ledger, error) {
	// Load the company (with its gateway assignments) tenant-scoped — this is a
	// supporting lookup for the engine, so it uses tenant-only scoping rather
	// than the owner-column scope, matching the transactions handler. A company
	// absent from the tenant leaves the engine on its defaults (0% gateway
	// commission), exactly as calc.js with an undefined company.
	var company *model.Company
	var c model.Company
	err := s.db.Preload("Gateways").
		Where("tenant_id = ? AND id = ?", p.TenantID, companyID).
		First(&c).Error
	switch {
	case err == nil:
		company = &c
	case errors.Is(err, gorm.ErrRecordNotFound):
		// leave company nil; engine defaults apply
	default:
		return Ledger{}, err
	}

	// Receivable: Σ companyNetIncome over the company's transactions. The
	// aggregation query carries the owner scope (transactions.company_id), so a
	// portal Company principal only ever sums its own company's transactions.
	var txns []model.Transaction
	if err := s.scoped(p).Where("company_id = ?", companyID).Find(&txns).Error; err != nil {
		return Ledger{}, err
	}
	var receivable float64
	for i := range txns {
		bd := commission.Calc(txns[i], commission.TxnContext{Company: company})
		receivable += bd.CompanyNetIncome
	}

	// Paid: Σ settlements to the company.
	paid, err := s.sumSettlements(p, companyID)
	if err != nil {
		return Ledger{}, err
	}

	return Ledger{Receivable: receivable, Paid: paid, Balance: receivable - paid}, nil
}

// AffiliateLedger computes the ledger for the affiliate with the given id
// within the principal's scope (Req 10.2). Earned is the sum of beneficiary
// commission across the transactions of the affiliate's merchants; paid is the
// sum of affiliate payments; balance is earned minus paid.
func (s *Service) AffiliateLedger(p service.Principal, affiliateID string) (Ledger, error) {
	// Affiliate record (tenant-scoped supporting lookup) supplies the
	// commission percentage and base the engine applies.
	var affiliate *model.Affiliate
	var a model.Affiliate
	err := s.db.Where("tenant_id = ? AND id = ?", p.TenantID, affiliateID).First(&a).Error
	switch {
	case err == nil:
		affiliate = &a
	case errors.Is(err, gorm.ErrRecordNotFound):
		// No such affiliate in this tenant: earned stays 0; only payments (also
		// scoped, so also empty) are summed.
	default:
		return Ledger{}, err
	}

	// The affiliate's merchants, read through the owner scope
	// (merchants.affiliate_id) so a portal Affiliate principal only sees its
	// own merchants. Keyed by id for the per-transaction engine context.
	var merchants []model.Merchant
	if err := s.scoped(p).Where("affiliate_id = ?", affiliateID).Find(&merchants).Error; err != nil {
		return Ledger{}, err
	}
	merchantByID := make(map[string]*model.Merchant, len(merchants))
	merchantIDs := make([]string, 0, len(merchants))
	for i := range merchants {
		merchantByID[merchants[i].ID] = &merchants[i]
		merchantIDs = append(merchantIDs, merchants[i].ID)
	}

	var earned float64
	if affiliate != nil && len(merchantIDs) > 0 {
		// Transactions of those merchants. Ownership is already established by
		// the affiliate-scoped merchant set above; the transactions table has
		// no affiliate_id column, so this query is tenant-scoped (not
		// owner-scoped) and filtered to the resolved merchant ids.
		var txns []model.Transaction
		if err := s.db.Where("tenant_id = ? AND merchant_id IN ?", p.TenantID, merchantIDs).
			Find(&txns).Error; err != nil {
			return Ledger{}, err
		}
		for i := range txns {
			m := merchantByID[txns[i].MerchantID]
			bd := commission.Calc(txns[i], commission.TxnContext{Merchant: m, Affiliate: affiliate})
			// The affiliate keeps its gross cut LESS the merchant's own share,
			// which is carved out of it (AffiliateCommission), not the gross
			// BeneficiaryCommission.
			earned += bd.AffiliateCommission
		}
	}

	// Paid: Σ affiliate payments to the affiliate.
	var payments []model.AffiliatePayment
	if err := s.scoped(p).Where("affiliate_id = ?", affiliateID).Find(&payments).Error; err != nil {
		return Ledger{}, err
	}
	var paid float64
	for i := range payments {
		paid += payments[i].Amount
	}

	return Ledger{Earned: earned, Paid: paid, Balance: earned - paid}, nil
}

// MerchantLedger computes the ledger for the merchant with the given id within
// the principal's scope (Req 10.3, 10.4). Earned is the sum of beneficiary
// commission across the merchant's transactions, computed ONLY for a direct
// merchant (no affiliate assignment); for an affiliate-assigned merchant earned
// stays 0, mirroring calc.js merchantLedger (Req 10.4). Paid is the sum of
// merchant payments; balance is earned minus paid.
func (s *Service) MerchantLedger(p service.Principal, merchantID string) (Ledger, error) {
	// Merchant record (tenant-scoped supporting lookup) supplies the commission
	// percentage/base and whether the merchant is direct or affiliate-assigned.
	var merchant *model.Merchant
	var m model.Merchant
	err := s.db.Where("tenant_id = ? AND id = ?", p.TenantID, merchantID).First(&m).Error
	switch {
	case err == nil:
		merchant = &m
	case errors.Is(err, gorm.ErrRecordNotFound):
		// No such merchant in this tenant: earned stays 0; only payments are
		// summed (also scoped, so also empty).
	default:
		return Ledger{}, err
	}

	// An affiliate-assigned merchant still earns its own commission; it is
	// carved OUT OF the affiliate's cut rather than added on top, so the
	// affiliate record must be loaded for the engine to apply the cap.
	var affiliate *model.Affiliate
	if merchant != nil && merchant.AffiliateID != nil && *merchant.AffiliateID != "" {
		var a model.Affiliate
		aerr := s.db.Where("tenant_id = ? AND id = ?", p.TenantID, *merchant.AffiliateID).
			First(&a).Error
		switch {
		case aerr == nil:
			affiliate = &a
		case errors.Is(aerr, gorm.ErrRecordNotFound):
			// leave affiliate nil; the merchant is treated as direct
		default:
			return Ledger{}, aerr
		}
	}

	var earned float64
	if merchant != nil {
		var txns []model.Transaction
		if err := s.scoped(p).Where("merchant_id = ?", merchantID).Find(&txns).Error; err != nil {
			return Ledger{}, err
		}
		for i := range txns {
			bd := commission.Calc(txns[i], commission.TxnContext{
				Merchant:  merchant,
				Affiliate: affiliate,
			})
			earned += bd.MerchantCommission
		}
	}

	// Paid: Σ merchant payments to the merchant.
	var payments []model.MerchantPayment
	if err := s.scoped(p).Where("merchant_id = ?", merchantID).Find(&payments).Error; err != nil {
		return Ledger{}, err
	}
	var paid float64
	for i := range payments {
		paid += payments[i].Amount
	}

	return Ledger{Earned: earned, Paid: paid, Balance: earned - paid}, nil
}

// sumSettlements returns the total of the company's settlement amounts within
// the principal's scope.
func (s *Service) sumSettlements(p service.Principal, companyID string) (float64, error) {
	var settlements []model.Settlement
	if err := s.scoped(p).Where("company_id = ?", companyID).Find(&settlements).Error; err != nil {
		return 0, err
	}
	var paid float64
	for i := range settlements {
		paid += settlements[i].Amount
	}
	return paid, nil
}
