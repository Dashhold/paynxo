// Package commission implements the Commission_Engine (Req 9): a pure,
// deterministic port of the frontend calc.js commission calculation. It has no
// I/O and no dependencies beyond the data model, which makes it directly
// importable by the transactions handler and the report/ledger service, and
// trivially testable (including the calc.js equivalence property in task 9.2).
//
// The algorithm mirrors calc.js calcTransaction exactly so the engine produces
// results equivalent to the current frontend computation for the same inputs
// (Req 9.7).
package commission

import "pgcs/backend/internal/model"

// Charge-bearer values. When no gateway assignment exists for the transaction's
// gateway, the bearer defaults to Admin, matching calc.js.
const (
	ChargeBearerAdmin   = "Admin"
	ChargeBearerCompany = "Company"
)

// Beneficiary values: who receives the merchant/affiliate commission.
const (
	BeneficiaryMerchant  = "Merchant"
	BeneficiaryAffiliate = "Affiliate"
)

// Commission-base values: the amount the beneficiary percentage is applied to.
const (
	BaseTransactionAmount = "Transaction Amount"
	BaseSettlementAmount  = "Settlement Amount"
)

// TxnContext supplies the related records the engine needs to compute a
// breakdown. Affiliate is nil when the merchant has no affiliate. Any field may
// be nil; the engine degrades gracefully exactly as calc.js does with optional
// chaining (e.g. a nil Company yields a 0% gateway commission and an Admin
// charge bearer).
type TxnContext struct {
	Company   *model.Company   // with its Gateways assignments; may be nil
	Merchant  *model.Merchant  // may be nil
	Affiliate *model.Affiliate // nil when the merchant has no affiliate
}

// Breakdown is the full per-transaction commission breakdown. The field set
// mirrors the object returned by calc.js calcTransaction, including the
// ChargesDeducted alias kept for backward compatibility.
type Breakdown struct {
	TxnAmount        float64 `json:"txnAmount"`
	SettlementAmount float64 `json:"settlementAmount"`
	TxnCharges       float64 `json:"txnCharges"`
	OtherCharges     float64 `json:"otherCharges"`

	GatewayCommissionPct float64 `json:"gatewayCommissionPct"`
	ChargeBearer         string  `json:"chargeBearer"` // Admin | Company
	AdminCommission      float64 `json:"adminCommission"`

	// Beneficiary describes the primary commission recipient and the TOTAL
	// commission the admin pays out for this transaction. When the merchant is
	// affiliate-assigned the beneficiary is the Affiliate and
	// BeneficiaryCommission is the affiliate's gross cut, of which the
	// merchant's own share is then carved out (see MerchantCommission /
	// AffiliateCommission). The total paid out of the admin's commission is
	// always BeneficiaryCommission.
	Beneficiary           string  `json:"beneficiary"`     // Merchant | Affiliate
	BeneficiaryPct        float64 `json:"beneficiaryPct"`
	BeneficiaryBase       string  `json:"beneficiaryBase"` // Transaction Amount | Settlement Amount
	BeneficiaryCommission float64 `json:"beneficiaryCommission"`

	// MerchantCommission is the merchant's own commission. For a direct
	// merchant it equals BeneficiaryCommission. For an affiliate-assigned
	// merchant it is carved OUT OF the affiliate's gross cut (capped at it), so
	// the admin's total payout is unchanged and the affiliate's net is reduced
	// by exactly this amount.
	MerchantPct        float64 `json:"merchantPct"`
	MerchantBase       string  `json:"merchantBase"`
	MerchantCommission float64 `json:"merchantCommission"`

	// AffiliateGrossCommission is the affiliate's cut before the merchant's
	// share is carved out; AffiliateCommission is what the affiliate actually
	// keeps. Both are 0 for a direct merchant.
	AffiliatePct             float64 `json:"affiliatePct"`
	AffiliateBase            string  `json:"affiliateBase"`
	AffiliateGrossCommission float64 `json:"affiliateGrossCommission"`
	AffiliateCommission      float64 `json:"affiliateCommission"`

	// ChargesDeducted is kept for backward compatibility with calc.js; it
	// equals CompanyChargesDeducted.
	ChargesDeducted        float64 `json:"chargesDeducted"`
	CompanyChargesDeducted float64 `json:"companyChargesDeducted"`
	AdminChargesDeducted   float64 `json:"adminChargesDeducted"`

	AdminNetCommission float64 `json:"adminNetCommission"`
	CompanyNetIncome   float64 `json:"companyNetIncome"`
}

// Calc computes the commission breakdown for a single transaction. It is a pure
// function: it performs no I/O, mutates nothing, and is fully deterministic.
//
// The steps follow calc.js calcTransaction exactly:
//  1. Find the company's gateway assignment matching txn.GatewayID; take the
//     gateway commission percentage and charge bearer from it (defaults: 0% and
//     "Admin" when no assignment is present) (Req 9.1).
//  2. Admin commission = txnAmount * gatewayCommissionPct / 100 (Req 9.1).
//  3. Merchant commission = base * merchant.CommissionPct / 100, always
//     computed from the merchant's own percentage/base (Req 9.5).
//  4. When the merchant is assigned to an affiliate and an affiliate is
//     present, the affiliate is the beneficiary: its percentage/base yields the
//     GROSS cut, and the merchant's commission is carved out of that gross
//     (capped at it) so the affiliate keeps gross - merchant (Req 9.4). The
//     total the admin pays out (beneficiaryCommission) is the gross either way.
//     Base selection: "Transaction Amount" uses txnAmount, otherwise
//     settlementAmount (Req 9.6).
//  5. Company net income = settlementAmount - adminCommission, less txn+other
//     charges when the charge bearer is Company (Req 9.2).
//  6. Admin net commission = adminCommission - beneficiaryCommission, less
//     txn+other charges when the charge bearer is Admin (Req 9.3).
func Calc(txn model.Transaction, ctx TxnContext) Breakdown {
	txnAmount := txn.TxnAmount
	settlementAmount := txn.SettlementAmount
	txnCharges := txn.TxnCharges
	otherCharges := txn.OtherCharges

	// 1. Gateway commission config from the company assignment (calc.js gwAssign).
	gatewayCommissionPct := 0.0
	chargeBearer := ChargeBearerAdmin
	if ctx.Company != nil {
		for i := range ctx.Company.Gateways {
			if ctx.Company.Gateways[i].GatewayID == txn.GatewayID {
				gatewayCommissionPct = ctx.Company.Gateways[i].Commission
				chargeBearer = ctx.Company.Gateways[i].ChargeBearer
				break
			}
		}
	}

	// 2. Admin commission = gateway % on the TRANSACTION amount.
	adminCommission := txnAmount * gatewayCommissionPct / 100

	// 3. Merchant / Affiliate commission.
	//
	// The merchant's own percentage/base always yields a merchant commission.
	// When the merchant is affiliate-assigned, the affiliate's percentage/base
	// yields the GROSS beneficiary cut and the merchant's share is carved OUT
	// of it (capped at the gross so the affiliate's net never goes negative).
	// The admin therefore pays the same total either way.
	baseFor := func(b string) float64 {
		if b == BaseTransactionAmount {
			return txnAmount
		}
		return settlementAmount
	}

	merchantPct := 0.0
	merchantBase := BaseSettlementAmount
	if ctx.Merchant != nil {
		merchantPct = ctx.Merchant.CommissionPct
		merchantBase = orDefault(ctx.Merchant.CommissionBase, BaseSettlementAmount)
	}
	merchantCommission := baseFor(merchantBase) * merchantPct / 100

	affiliatePct := 0.0
	affiliateBase := ""
	affiliateGrossCommission := 0.0
	affiliateCommission := 0.0

	beneficiary := BeneficiaryMerchant
	beneficiaryPct := merchantPct
	beneficiaryBase := merchantBase
	beneficiaryCommission := merchantCommission

	if ctx.Merchant != nil && ctx.Merchant.AffiliateID != nil &&
		*ctx.Merchant.AffiliateID != "" && ctx.Affiliate != nil {
		beneficiary = BeneficiaryAffiliate
		affiliatePct = ctx.Affiliate.CommissionPct
		affiliateBase = orDefault(ctx.Affiliate.CommissionBase, BaseSettlementAmount)
		affiliateGrossCommission = baseFor(affiliateBase) * affiliatePct / 100

		// Carve the merchant's share out of the affiliate's gross cut.
		if merchantCommission > affiliateGrossCommission {
			merchantCommission = affiliateGrossCommission
		}
		affiliateCommission = affiliateGrossCommission - merchantCommission

		// The total the admin pays out stays the affiliate's gross cut.
		beneficiaryPct = affiliatePct
		beneficiaryBase = affiliateBase
		beneficiaryCommission = affiliateGrossCommission
	}

	// 5. Company net income. Admin earns the gateway commission FROM the
	//    company; the merchant/affiliate commission is NOT charged to the
	//    company. Charges are deducted from whoever is the charge bearer.
	companyChargesDeducted := 0.0
	companyNetIncome := settlementAmount - adminCommission
	if chargeBearer == ChargeBearerCompany {
		companyChargesDeducted = txnCharges + otherCharges
		companyNetIncome -= companyChargesDeducted
	}

	// 6. Admin net commission. The merchant/affiliate commission is paid OUT OF
	//    the admin's commission.
	adminChargesDeducted := 0.0
	adminNetCommission := adminCommission - beneficiaryCommission
	if chargeBearer == ChargeBearerAdmin {
		adminChargesDeducted = txnCharges + otherCharges
		adminNetCommission -= adminChargesDeducted
	}

	return Breakdown{
		TxnAmount:            txnAmount,
		SettlementAmount:     settlementAmount,
		TxnCharges:           txnCharges,
		OtherCharges:         otherCharges,
		GatewayCommissionPct: gatewayCommissionPct,
		ChargeBearer:         chargeBearer,
		AdminCommission:      adminCommission,

		Beneficiary:           beneficiary,
		BeneficiaryPct:        beneficiaryPct,
		BeneficiaryBase:       beneficiaryBase,
		BeneficiaryCommission: beneficiaryCommission,

		MerchantPct:        merchantPct,
		MerchantBase:       merchantBase,
		MerchantCommission: merchantCommission,

		AffiliatePct:             affiliatePct,
		AffiliateBase:            affiliateBase,
		AffiliateGrossCommission: affiliateGrossCommission,
		AffiliateCommission:      affiliateCommission,

		ChargesDeducted:        companyChargesDeducted, // backward-compat alias
		CompanyChargesDeducted: companyChargesDeducted,
		AdminChargesDeducted:   adminChargesDeducted,

		AdminNetCommission: adminNetCommission,
		CompanyNetIncome:   companyNetIncome,
	}
}

// orDefault returns v unless it is empty, in which case it returns def. This
// mirrors the JavaScript `value || default` idiom calc.js uses for the
// commission base.
func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
