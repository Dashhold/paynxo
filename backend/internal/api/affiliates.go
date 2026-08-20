package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// affiliateHandlers serves the /api/affiliates CRUD endpoints. Affiliates carry
// a portal login account, so the handlers delegate to the account-aware
// AffiliateService.
type affiliateHandlers struct {
	svc *crud.AffiliateService
}

// newAffiliateHandlers builds the affiliate handlers over a database handle.
func newAffiliateHandlers(deps Deps) *affiliateHandlers {
	return &affiliateHandlers{svc: crud.NewAffiliateService(deps.DB)}
}

// list handles GET /api/affiliates within the principal's tenant scope
// (Req 8.5).
func (h *affiliateHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/affiliates. The tenant is assigned from the
// principal (Req 4.2); a missing name yields a 400 identifying the field
// (Req 8.4).
func (h *affiliateHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var a model.Affiliate
	if err := decodeJSON(r, &a); err != nil {
		return err
	}
	if err := h.svc.Create(p, &a); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, a)
	return nil
}

// get handles GET /api/affiliates/{id}; out-of-scope yields 404 (Req 4.3).
func (h *affiliateHandlers) get(w http.ResponseWriter, r *http.Request) error {
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

// update handles PUT /api/affiliates/{id}. The path id is authoritative; an
// out-of-scope record yields 404 (Req 4.3).
func (h *affiliateHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var a model.Affiliate
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

// del handles DELETE /api/affiliates/{id}, returning 204 or 404 (Req 4.3).
func (h *affiliateHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
