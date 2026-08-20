package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// bankHandlers serves the /api/banks CRUD endpoints. Banks are simple
// tenant-scoped reference entities with no nested data, so the handlers
// delegate to the generic tenant-scoped CRUD service following the same thin
// pattern as the gateway handlers.
type bankHandlers struct {
	svc *crud.Service[model.Bank, *model.Bank]
}

// newBankHandlers builds the bank handlers over a database handle.
func newBankHandlers(deps Deps) *bankHandlers {
	return &bankHandlers{
		svc: crud.NewService[model.Bank](deps.DB, crud.ValidateBank),
	}
}

// list handles GET /api/banks within the principal's tenant scope (Req 8.5).
func (h *bankHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/banks. The tenant is assigned from the principal
// (Req 4.2); a missing name/code yields a 400 identifying the field (Req 8.4).
func (h *bankHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var b model.Bank
	if err := decodeJSON(r, &b); err != nil {
		return err
	}
	if err := h.svc.Create(p, &b); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, b)
	return nil
}

// get handles GET /api/banks/{id}; out-of-scope yields 404 (Req 4.3).
func (h *bankHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	b, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, b)
	return nil
}

// update handles PUT /api/banks/{id}. The path id is authoritative; an
// out-of-scope record yields 404 (Req 4.3).
func (h *bankHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var b model.Bank
	if err := decodeJSON(r, &b); err != nil {
		return err
	}
	b.ID = r.PathValue("id")
	if err := h.svc.Update(p, &b); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, b)
	return nil
}

// del handles DELETE /api/banks/{id}, returning 204 or 404 (Req 4.3).
func (h *bankHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
