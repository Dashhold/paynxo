package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/crud"
)

// companyHandlers serves the /api/companies CRUD endpoints. Companies carry
// nested CompanyGateway assignments, so the handlers delegate to the
// association-aware CompanyService: reads return the assignments, creates
// persist them, and updates fully replace them (Req 8.2).
type companyHandlers struct {
	svc *crud.CompanyService
}

// newCompanyHandlers builds the company handlers over a database handle.
func newCompanyHandlers(deps Deps) *companyHandlers {
	return &companyHandlers{svc: crud.NewCompanyService(deps.DB)}
}

// list handles GET /api/companies within the principal's tenant scope, each
// company including its gateway assignments (Req 8.5).
func (h *companyHandlers) list(w http.ResponseWriter, r *http.Request) error {
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

// create handles POST /api/companies, persisting the company and its gateway
// assignments together (Req 8.2). The tenant is assigned from the principal
// (Req 4.2); a missing name yields a 400 identifying the field (Req 8.4).
func (h *companyHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var c model.Company
	if err := decodeJSON(r, &c); err != nil {
		return err
	}
	if err := h.svc.Create(p, &c); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusCreated, c)
	return nil
}

// get handles GET /api/companies/{id} with its gateway assignments preloaded;
// out-of-scope yields 404 (Req 4.3, 18.2).
func (h *companyHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	c, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, c)
	return nil
}

// update handles PUT /api/companies/{id}, fully replacing the company's gateway
// assignments with the supplied set (Req 8.2). The path id is authoritative; an
// out-of-scope record yields 404 (Req 4.3).
func (h *companyHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var c model.Company
	if err := decodeJSON(r, &c); err != nil {
		return err
	}
	c.ID = r.PathValue("id")
	if err := h.svc.Update(p, &c); err != nil {
		return err
	}
	middleware.WriteJSON(w, http.StatusOK, c)
	return nil
}

// del handles DELETE /api/companies/{id}, removing the company and its gateway
// assignments; an out-of-scope record yields 404 (Req 4.3).
func (h *companyHandlers) del(w http.ResponseWriter, r *http.Request) error {
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
