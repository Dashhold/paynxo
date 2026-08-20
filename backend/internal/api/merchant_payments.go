package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// merchantPaymentHandlers serves the /api/merchant-payments CRUD endpoints.
// Merchant payments are commission payouts to a merchant; they carry no nested
// data, so the handlers delegate to the generic tenant-scoped CRUD service
// following the same thin pattern as the other entity handlers.
type merchantPaymentHandlers struct {
	svc *crud.Service[model.MerchantPayment, *model.MerchantPayment]
}

// newMerchantPaymentHandlers builds the merchant-payment handlers over a
// database handle.
func newMerchantPaymentHandlers(deps Deps) *merchantPaymentHandlers {
	return &merchantPaymentHandlers{
		svc: crud.NewService[model.MerchantPayment](deps.DB, crud.ValidateMerchantPayment),
	}
}

// list handles GET /api/merchant-payments within the principal's tenant scope
// (Req 8.5).
func (h *merchantPaymentHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/merchant-payments. The tenant is assigned from the
// principal (Req 4.2); missing required fields yield a 400 identifying the
// field (Req 8.4).
func (h *merchantPaymentHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var m model.MerchantPayment
	if err := decodeJSON(r, &m); err != nil {
		return err
	}
	if err := h.svc.Create(p, &m); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, m)
	return nil
}

// get handles GET /api/merchant-payments/{id}; out-of-scope yields 404
// (Req 4.3).
func (h *merchantPaymentHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	m, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, m)
	return nil
}

// update handles PUT /api/merchant-payments/{id}. The path id is authoritative;
// an out-of-scope record yields 404 (Req 4.3).
func (h *merchantPaymentHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var m model.MerchantPayment
	if err := decodeJSON(r, &m); err != nil {
		return err
	}
	m.ID = r.PathValue("id")
	if err := h.svc.Update(p, &m); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, m)
	return nil
}

// del handles DELETE /api/merchant-payments/{id}, returning 204 or 404
// (Req 4.3).
func (h *merchantPaymentHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
