package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// companyPaymentHandlers serves the /api/company-payments endpoints. A payment
// to a company is modelled as a Settlement (a payment made to a company), so
// these handlers reuse the generic tenant-scoped Settlement CRUD service. This
// gives the mobile app a dedicated, intention-revealing endpoint for recording
// company payments while persisting through the existing settlement store.
type companyPaymentHandlers struct {
	svc *crud.Service[model.Settlement, *model.Settlement]
}

// newCompanyPaymentHandlers builds the company-payment handlers over a database
// handle.
func newCompanyPaymentHandlers(deps Deps) *companyPaymentHandlers {
	return &companyPaymentHandlers{
		svc: crud.NewService[model.Settlement](deps.DB, crud.ValidateSettlement),
	}
}

// list handles GET /api/company-payments within the principal's tenant scope
// (Req 8.5).
func (h *companyPaymentHandlers) list(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	items, err := h.svc.List(p)
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// create handles POST /api/company-payments. The tenant is assigned from the
// principal (Req 4.2); a missing company or non-positive amount yields a 400
// identifying the field (Req 8.4).
func (h *companyPaymentHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var s model.Settlement
	if err := decodeJSON(r, &s); err != nil {
		return err
	}
	if err := h.svc.Create(p, &s); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, s)
	return nil
}

// get handles GET /api/company-payments/{id}; out-of-scope yields 404 (Req 4.3).
func (h *companyPaymentHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	s, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, s)
	return nil
}

// del handles DELETE /api/company-payments/{id}, returning 204 or 404 (Req 4.3).
func (h *companyPaymentHandlers) del(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(p, r.PathValue("id")); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
