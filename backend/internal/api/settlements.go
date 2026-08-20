package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// settlementHandlers serves the /api/settlements CRUD endpoints. Settlements
// are payments made to a company; they carry no nested data, so the handlers
// delegate to the generic tenant-scoped CRUD service, following the same thin
// pattern as the gateway and affiliate handlers.
type settlementHandlers struct {
	svc *crud.Service[model.Settlement, *model.Settlement]
}

// newSettlementHandlers builds the settlement handlers over a database handle.
func newSettlementHandlers(deps Deps) *settlementHandlers {
	return &settlementHandlers{
		svc: crud.NewService[model.Settlement](deps.DB, crud.ValidateSettlement),
	}
}

// list handles GET /api/settlements within the principal's tenant scope
// (Req 8.5).
func (h *settlementHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/settlements. The tenant is assigned from the
// principal (Req 4.2); missing required fields yield a 400 identifying the
// field (Req 8.4).
func (h *settlementHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var s model.Settlement
	if err := decodeJSON(r, &s); err != nil {
		return err
	}
	// The recording role is taken from the principal, never the request body,
	// so a company-recorded settlement stays distinguishable from this one.
	s.RecordedByRole = p.Role
	if err := h.svc.Create(p, &s); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, s)
	return nil
}

// get handles GET /api/settlements/{id}; out-of-scope yields 404 (Req 4.3).
func (h *settlementHandlers) get(w http.ResponseWriter, r *http.Request) error {
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

// update handles PUT /api/settlements/{id}. The path id is authoritative; an
// out-of-scope record yields 404 (Req 4.3).
func (h *settlementHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var s model.Settlement
	if err := decodeJSON(r, &s); err != nil {
		return err
	}
	s.ID = r.PathValue("id")
	if err := h.svc.Update(p, &s); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, s)
	return nil
}

// del handles DELETE /api/settlements/{id}, returning 204 or 404 (Req 4.3).
func (h *settlementHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
