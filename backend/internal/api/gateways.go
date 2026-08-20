package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// gatewayHandlers serves the /api/gateways CRUD endpoints. Like the auth
// handlers, each method is thin: it recovers the tenant-scoped principal,
// decodes the request, delegates to the CRUD service, and encodes the result,
// propagating typed errors for the Error middleware to render.
type gatewayHandlers struct {
	svc *crud.Service[model.Gateway, *model.Gateway]
}

// newGatewayHandlers builds the gateway handlers over a database handle.
func newGatewayHandlers(deps Deps) *gatewayHandlers {
	return &gatewayHandlers{
		svc: crud.NewService[model.Gateway](deps.DB, crud.ValidateGateway),
	}
}

// list handles GET /api/gateways: every gateway within the principal's tenant
// scope (Req 8.5).
func (h *gatewayHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/gateways. The tenant is assigned from the principal,
// never the request body (Req 4.2); a missing name yields a 400 identifying the
// field (Req 8.4).
func (h *gatewayHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var g model.Gateway
	if err := decodeJSON(r, &g); err != nil {
		return err
	}
	if err := h.svc.Create(p, &g); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, g)
	return nil
}

// get handles GET /api/gateways/{id}. A gateway outside the principal's scope
// is reported as 404 (Req 4.3, 18.2).
func (h *gatewayHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	g, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, g)
	return nil
}

// update handles PUT /api/gateways/{id}. The path id is authoritative, and
// updating a gateway outside the principal's scope returns 404 (Req 4.3).
func (h *gatewayHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var g model.Gateway
	if err := decodeJSON(r, &g); err != nil {
		return err
	}
	g.ID = r.PathValue("id")
	if err := h.svc.Update(p, &g); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, g)
	return nil
}

// del handles DELETE /api/gateways/{id}, returning 204 on success or 404 when
// no in-scope gateway matches (Req 4.3).
func (h *gatewayHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
