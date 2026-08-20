package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// affiliatePaymentHandlers serves the /api/affiliate-payments CRUD endpoints.
// Affiliate payments are commission payouts to an affiliate; they carry no
// nested data, so the handlers delegate to the generic tenant-scoped CRUD
// service following the same thin pattern as the other entity handlers.
type affiliatePaymentHandlers struct {
	svc *crud.Service[model.AffiliatePayment, *model.AffiliatePayment]
}

// newAffiliatePaymentHandlers builds the affiliate-payment handlers over a
// database handle.
func newAffiliatePaymentHandlers(deps Deps) *affiliatePaymentHandlers {
	return &affiliatePaymentHandlers{
		svc: crud.NewService[model.AffiliatePayment](deps.DB, crud.ValidateAffiliatePayment),
	}
}

// list handles GET /api/affiliate-payments within the principal's tenant scope
// (Req 8.5).
func (h *affiliatePaymentHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/affiliate-payments. The tenant is assigned from the
// principal (Req 4.2); missing required fields yield a 400 identifying the
// field (Req 8.4).
func (h *affiliatePaymentHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var a model.AffiliatePayment
	if err := decodeJSON(r, &a); err != nil {
		return err
	}
	if err := h.svc.Create(p, &a); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, a)
	return nil
}

// get handles GET /api/affiliate-payments/{id}; out-of-scope yields 404
// (Req 4.3).
func (h *affiliatePaymentHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	a, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, a)
	return nil
}

// update handles PUT /api/affiliate-payments/{id}. The path id is
// authoritative; an out-of-scope record yields 404 (Req 4.3).
func (h *affiliatePaymentHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var a model.AffiliatePayment
	if err := decodeJSON(r, &a); err != nil {
		return err
	}
	a.ID = r.PathValue("id")
	if err := h.svc.Update(p, &a); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, a)
	return nil
}

// del handles DELETE /api/affiliate-payments/{id}, returning 204 or 404
// (Req 4.3).
func (h *affiliatePaymentHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
