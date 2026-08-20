package api

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/commission"
	"pgcs/backend/internal/service/crud"
)

// portalHandlers serves the read-only /api/portal/* endpoints used by the
// Company, Affiliate, and Merchant portals.
//
// The core entity endpoints (/api/merchants, /api/transactions, ...) are
// Admin/SuperAdmin only. Portal principals nevertheless need to read the slice
// of data they own (their merchants, transactions, settlements) so their
// dashboards and lists can render — exactly the data the web portal pages show.
//
// These handlers provide that, scoped explicitly per role to the records the
// principal owns within its tenant. Transaction reads embed the same commission
// breakdown as /api/transactions, and merchant reads reuse
// redactMerchantSecrets so a portal principal only ever sees the secrets of its
// own merchant record.
//
// Reads remain available to all three portal roles. Writes are additionally
// available to the Company role only, and only for its OWN transactions and
// settlements: the write handlers below overwrite the companyId on the incoming
// payload with the principal's OwnerID, and the underlying services resolve
// existing records through repo.ScopeTenant (which pins a Company principal to
// company_id = OwnerID), so a Company can never create or modify a record
// belonging to another company.
type portalHandlers struct {
	deps Deps
	txns *crud.TransactionService
	stls *crud.Service[model.Settlement, *model.Settlement]
}

// newPortalHandlers builds the portal handlers over the shared deps.
func newPortalHandlers(deps Deps) *portalHandlers {
	return &portalHandlers{
		deps: deps,
		txns: crud.NewTransactionService(deps.DB),
		stls: crud.NewService[model.Settlement](deps.DB, crud.ValidateSettlement),
	}
}

// merchants handles GET /api/portal/merchants. It returns the merchants the
// portal principal owns, with their nested tree preloaded and secrets redacted
// for the caller:
//   - Company   -> merchants where company_id = OwnerID  ("My Merchants")
//   - Affiliate -> merchants where affiliate_id = OwnerID ("My Merchants")
//   - Merchant  -> the principal's own merchant record (for "My Banks")
func (h *portalHandlers) merchants(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}

	q := h.deps.DB.
		Preload("Banks.AtmCards").
		Preload("Banks.Custom").
		Preload("PaymentGateways.Custom").
		Where("tenant_id = ?", p.TenantID)

	switch p.Role {
	case service.RoleCompany:
		q = q.Where("company_id = ?", p.OwnerID)
	case service.RoleAffiliate:
		q = q.Where("affiliate_id = ?", p.OwnerID)
	case service.RoleMerchant:
		q = q.Where("id = ?", p.OwnerID)
	default:
		// No other role reaches this route (guarded by RequireRoles), but fail
		// closed with an empty result rather than disclosing the whole tenant.
		middleware.WriteJSON(w, http.StatusOK, []model.Merchant{})
		return nil
	}

	var items []model.Merchant
	if err := q.Find(&items).Error; err != nil {
		return err
	}
	for i := range items {
		redactMerchantSecrets(p, &items[i])
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// transactions handles GET /api/portal/transactions. It returns the
// transactions the portal principal owns, each enriched with its computed
// commission breakdown (the same shape as /api/transactions):
//   - Company   -> transactions where company_id = OwnerID
//   - Affiliate -> transactions of the affiliate's merchants
//   - Merchant  -> transactions where merchant_id = OwnerID
func (h *portalHandlers) transactions(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}

	var txns []model.Transaction
	switch p.Role {
	case service.RoleCompany:
		if err := h.deps.DB.
			Where("tenant_id = ? AND company_id = ?", p.TenantID, p.OwnerID).
			Find(&txns).Error; err != nil {
			return err
		}
	case service.RoleMerchant:
		if err := h.deps.DB.
			Where("tenant_id = ? AND merchant_id = ?", p.TenantID, p.OwnerID).
			Find(&txns).Error; err != nil {
			return err
		}
	case service.RoleAffiliate:
		// The transactions table carries no affiliate_id, so resolve the
		// affiliate's merchants first (affiliate_id = OwnerID), then their
		// transactions — mirroring the affiliate ledger aggregation.
		var merchantIDs []string
		if err := h.deps.DB.Model(&model.Merchant{}).
			Where("tenant_id = ? AND affiliate_id = ?", p.TenantID, p.OwnerID).
			Pluck("id", &merchantIDs).Error; err != nil {
			return err
		}
		if len(merchantIDs) > 0 {
			if err := h.deps.DB.
				Where("tenant_id = ? AND merchant_id IN ?", p.TenantID, merchantIDs).
				Find(&txns).Error; err != nil {
				return err
			}
		}
	default:
		middleware.WriteJSON(w, http.StatusOK, []crud.TransactionWithBreakdown{})
		return nil
	}

	out := make([]crud.TransactionWithBreakdown, 0, len(txns))
	for i := range txns {
		bd, err := h.breakdown(p, txns[i])
		if err != nil {
			return err
		}
		out = append(out, crud.TransactionWithBreakdown{Transaction: txns[i], Breakdown: bd})
	}
	middleware.WriteJSON(w, http.StatusOK, out)
	return nil
}

// settlements handles GET /api/portal/settlements. Only the Company portal
// shows settlements; it returns the settlements paid to the principal's company
// (company_id = OwnerID).
func (h *portalHandlers) settlements(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	if p.Role != service.RoleCompany {
		middleware.WriteJSON(w, http.StatusOK, []model.Settlement{})
		return nil
	}
	var items []model.Settlement
	if err := h.deps.DB.
		Where("tenant_id = ? AND company_id = ?", p.TenantID, p.OwnerID).
		Find(&items).Error; err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// affiliatePayments handles GET /api/portal/affiliate-payments. Only the
// Affiliate portal uses it; it returns the commission payments made to the
// principal's affiliate (affiliate_id = OwnerID) so the Commission Ledger can
// render its running balance exactly like the web.
func (h *portalHandlers) affiliatePayments(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	if p.Role != service.RoleAffiliate {
		middleware.WriteJSON(w, http.StatusOK, []model.AffiliatePayment{})
		return nil
	}
	var items []model.AffiliatePayment
	if err := h.deps.DB.
		Where("tenant_id = ? AND affiliate_id = ?", p.TenantID, p.OwnerID).
		Find(&items).Error; err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// merchantPayments handles GET /api/portal/merchant-payments. Only the Merchant
// portal uses it; it returns the commission payments made to the principal's
// merchant (merchant_id = OwnerID).
func (h *portalHandlers) merchantPayments(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	if p.Role != service.RoleMerchant {
		middleware.WriteJSON(w, http.StatusOK, []model.MerchantPayment{})
		return nil
	}
	var items []model.MerchantPayment
	if err := h.deps.DB.
		Where("tenant_id = ? AND merchant_id = ?", p.TenantID, p.OwnerID).
		Find(&items).Error; err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// ---------------------------------------------------------------------------
// Company writes
//
// A Company records its own day-to-day activity: the transactions its merchants
// put through, and the settlements it receives. Both are pinned to the
// principal's own company so a Company can only ever write its own records.
// ---------------------------------------------------------------------------

// requireCompany resolves the principal and confirms it is a Company. Only the
// Company role reaches the portal write routes; any other role fails closed
// with a 403 even though RequireRoles already guards the mount point.
func requireCompany(r *http.Request) (service.Principal, error) {
	p, err := requirePrincipal(r)
	if err != nil {
		return p, err
	}
	if p.Role != service.RoleCompany {
		return p, apierr.ErrForbidden
	}
	return p, nil
}

// createTransaction handles POST /api/portal/transactions. The transaction's
// company is taken from the principal (never from the request body), so a
// Company can only record transactions against itself. Commissions are derived
// from the stored record on read, so the admin's commission and the
// merchant/affiliate split update as soon as the transaction lands.
func (h *portalHandlers) createTransaction(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	var t model.Transaction
	if err := decodeJSON(r, &t); err != nil {
		return err
	}
	t.CompanyID = p.OwnerID
	t.RecordedByRole = p.Role
	if err := h.txns.Create(p, &t); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, t)
	return nil
}

// updateTransaction handles PUT /api/portal/transactions/{id}. The path id and
// the principal's company are authoritative; a transaction belonging to another
// company is out of scope and yields 404 (Req 4.3).
func (h *portalHandlers) updateTransaction(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	var t model.Transaction
	if err := decodeJSON(r, &t); err != nil {
		return err
	}
	t.ID = r.PathValue("id")
	t.CompanyID = p.OwnerID
	t.RecordedByRole = p.Role
	if err := h.txns.Update(p, &t); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, t)
	return nil
}

// deleteTransaction handles DELETE /api/portal/transactions/{id}, returning 204
// or 404 for a transaction outside the principal's company (Req 4.3).
func (h *portalHandlers) deleteTransaction(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	if err := h.txns.Delete(p, r.PathValue("id")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// createSettlement handles POST /api/portal/settlements. The settlement's
// company is taken from the principal, and its Cash/Bank payment details are
// validated by crud.ValidateSettlement like an admin-recorded settlement.
func (h *portalHandlers) createSettlement(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	var s model.Settlement
	if err := decodeJSON(r, &s); err != nil {
		return err
	}
	s.CompanyID = p.OwnerID
	s.RecordedByRole = p.Role
	if err := h.stls.Create(p, &s); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, s)
	return nil
}

// updateSettlement handles PUT /api/portal/settlements/{id}; a settlement
// belonging to another company is out of scope and yields 404 (Req 4.3).
func (h *portalHandlers) updateSettlement(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	var s model.Settlement
	if err := decodeJSON(r, &s); err != nil {
		return err
	}
	s.ID = r.PathValue("id")
	s.CompanyID = p.OwnerID
	s.RecordedByRole = p.Role
	if err := h.stls.Update(p, &s); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, s)
	return nil
}

// deleteSettlement handles DELETE /api/portal/settlements/{id}, returning 204
// or 404 for a settlement outside the principal's company (Req 4.3).
func (h *portalHandlers) deleteSettlement(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	if err := h.stls.Delete(p, r.PathValue("id")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// banks handles GET /api/portal/banks. Recording a Bank payment requires
// choosing the bank the money was routed through, so the Company portal needs
// the tenant's bank list. It is a reference list with no secrets, and is
// returned tenant-scoped.
func (h *portalHandlers) banks(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var items []model.Bank
	if err := h.deps.DB.Where("tenant_id = ?", p.TenantID).Find(&items).Error; err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// gateways handles GET /api/portal/gateways. Recording a transaction requires
// choosing the payment gateway it ran through, so the Company portal needs the
// tenant's gateway list (a reference list with no secrets).
func (h *portalHandlers) gateways(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var items []model.Gateway
	if err := h.deps.DB.Where("tenant_id = ?", p.TenantID).Find(&items).Error; err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// company handles GET /api/portal/company. The Company portal needs its own
// company record — specifically its gateway assignments (commission percentage
// and charge bearer per gateway) — to preview the commission split while
// recording a transaction. Only the principal's own company is ever returned,
// and the portal login credentials are not part of the serialized shape.
func (h *portalHandlers) company(w http.ResponseWriter, r *http.Request) error {
	p, err := requireCompany(r)
	if err != nil {
		return err
	}
	var c model.Company
	err = h.deps.DB.Preload("Gateways").
		Where("tenant_id = ? AND id = ?", p.TenantID, p.OwnerID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierr.ErrNotFound
	}
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, []model.Company{c})
	return nil
}

// breakdown loads the related records a transaction needs and computes its
// commission breakdown, constrained to the principal's tenant. It mirrors the
// graceful-degradation behaviour of the admin transaction service: a missing
// related record leaves the engine on its defaults.
func (h *portalHandlers) breakdown(p service.Principal, txn model.Transaction) (commission.Breakdown, error) {
	var ctx commission.TxnContext

	if txn.CompanyID != "" {
		var c model.Company
		err := h.deps.DB.Preload("Gateways").
			Where("tenant_id = ? AND id = ?", p.TenantID, txn.CompanyID).
			First(&c).Error
		switch {
		case err == nil:
			ctx.Company = &c
		case errors.Is(err, gorm.ErrRecordNotFound):
			// leave Company nil; engine defaults apply
		default:
			return commission.Breakdown{}, err
		}
	}

	if txn.MerchantID != "" {
		var m model.Merchant
		err := h.deps.DB.
			Where("tenant_id = ? AND id = ?", p.TenantID, txn.MerchantID).
			First(&m).Error
		switch {
		case err == nil:
			ctx.Merchant = &m
			if m.AffiliateID != nil && *m.AffiliateID != "" {
				var af model.Affiliate
				aerr := h.deps.DB.
					Where("tenant_id = ? AND id = ?", p.TenantID, *m.AffiliateID).
					First(&af).Error
				switch {
				case aerr == nil:
					ctx.Affiliate = &af
				case errors.Is(aerr, gorm.ErrRecordNotFound):
					// leave Affiliate nil; merchant becomes the beneficiary
				default:
					return commission.Breakdown{}, aerr
				}
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			// leave Merchant nil; engine defaults apply
		default:
			return commission.Breakdown{}, err
		}
	}

	return commission.Calc(txn, ctx), nil
}
