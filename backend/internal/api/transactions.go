package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// transactionHandlers serves the /api/transactions CRUD endpoints. Unlike the
// simple entity handlers, every read returns the transaction together with its
// computed commission breakdown (Req 9): the service loads the related Company
// (with its gateway assignments), Merchant, and Affiliate within the tenant
// scope and runs the Commission_Engine, embedding the result alongside the
// transaction fields. Writes behave like the other tenant-scoped entities.
type transactionHandlers struct {
	svc *crud.TransactionService
}

// newTransactionHandlers builds the transaction handlers over a database
// handle.
func newTransactionHandlers(deps Deps) *transactionHandlers {
	return &transactionHandlers{svc: crud.NewTransactionService(deps.DB)}
}

// list handles GET /api/transactions within the principal's tenant scope, each
// transaction enriched with its computed commission breakdown (Req 8.5, 9).
func (h *transactionHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/transactions. The tenant is assigned from the
// principal (Req 4.2); missing required fields yield a 400 identifying the
// field (Req 8.4). The created transaction is returned without an enriched
// breakdown — clients read the breakdown via the GET endpoints.
func (h *transactionHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var t model.Transaction
	if err := decodeJSON(r, &t); err != nil {
		return err
	}
	// The recording role is taken from the principal, never the request body,
	// so a company-recorded transaction stays distinguishable from this one.
	t.RecordedByRole = p.Role
	if err := h.svc.Create(p, &t); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, t)
	return nil
}

// get handles GET /api/transactions/{id}, returning the transaction with its
// computed commission breakdown (Req 9); out-of-scope yields 404 (Req 4.3,
// 18.2).
func (h *transactionHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	t, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, t)
	return nil
}

// update handles PUT /api/transactions/{id}. The path id is authoritative; an
// out-of-scope record yields 404 (Req 4.3).
func (h *transactionHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var t model.Transaction
	if err := decodeJSON(r, &t); err != nil {
		return err
	}
	t.ID = r.PathValue("id")
	if err := h.svc.Update(p, &t); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, t)
	return nil
}

// del handles DELETE /api/transactions/{id}, returning 204 or 404 (Req 4.3).
func (h *transactionHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
