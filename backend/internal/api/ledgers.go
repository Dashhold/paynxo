package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/service/report"
)

// ledgerHandlers serves the read-only ledger endpoints (Req 10): the company,
// affiliate, and merchant ledgers computed by the Report_Service over
// tenant-scoped records. Every figure is aggregated through the principal's
// tenant + owner scope, so a portal principal only ever sees its own balances
// (Req 10.5).
type ledgerHandlers struct {
	svc *report.Service
}

// newLedgerHandlers builds the ledger handlers over a database handle.
func newLedgerHandlers(deps Deps) *ledgerHandlers {
	return &ledgerHandlers{svc: report.NewService(deps.DB)}
}

// company handles GET /api/ledgers/company/{id}, returning the company ledger
// (receivable/paid/balance) within the principal's scope (Req 10.1). An
// out-of-scope company simply yields zero balances rather than disclosing
// whether it exists in another tenant (Req 10.5).
func (h *ledgerHandlers) company(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	l, err := h.svc.CompanyLedger(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, l)
	return nil
}

// affiliate handles GET /api/ledgers/affiliate/{id}, returning the affiliate
// ledger (earned/paid/balance) within the principal's scope (Req 10.2).
func (h *ledgerHandlers) affiliate(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	l, err := h.svc.AffiliateLedger(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, l)
	return nil
}

// merchant handles GET /api/ledgers/merchant/{id}, returning the merchant
// ledger (earned/paid/balance) within the principal's scope. Only direct
// merchants earn commission; an affiliate-assigned merchant reports zero earned
// (Req 10.3, 10.4).
func (h *ledgerHandlers) merchant(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	l, err := h.svc.MerchantLedger(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, l)
	return nil
}
