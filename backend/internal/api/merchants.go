package api

import (
	"net/http"

	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/crud"
)

// merchantHandlers serves the /api/merchants CRUD endpoints. Merchants carry a
// deeply nested tree (banks -> ATM cards + custom fields; payment-gateway
// credentials -> custom fields), so the handlers delegate to the
// association-aware MerchantService: reads return the tree, creates and updates
// persist it together (Req 8.2), and deletes cascade it (Req 8.3).
//
// Before encoding a response, the handlers project the merchant through
// redactMerchantSecrets so secret-bearing fields are withheld from any
// principal that does not own the data (Req 8.7).
type merchantHandlers struct {
	svc *crud.MerchantService
}

// newMerchantHandlers builds the merchant handlers over a database handle.
func newMerchantHandlers(deps Deps) *merchantHandlers {
	return &merchantHandlers{svc: crud.NewMerchantService(deps.DB)}
}

// list handles GET /api/merchants within the principal's tenant scope, each
// merchant including its nested tree with secrets projected for the caller
// (Req 8.5, 8.7).
func (h *merchantHandlers) list(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	items, err := h.svc.List(p)
	if err != nil {
		return err
	}
	for i := range items {
		redactMerchantSecrets(p, &items[i])
	}
	middleware.WriteJSON(w, http.StatusOK, items)
	return nil
}

// create handles POST /api/merchants, persisting the merchant and its nested
// banks, ATM cards, payment-gateway credentials, and custom fields together
// (Req 8.2). The tenant is assigned from the principal (Req 4.2); a missing
// name yields a 400 identifying the field (Req 8.4).
func (h *merchantHandlers) create(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var m model.Merchant
	if err := decodeJSON(r, &m); err != nil {
		return err
	}
	if err := h.svc.Create(p, &m); err != nil {
		return err
	}
	redactMerchantSecrets(p, &m)
	middleware.WriteJSON(w, http.StatusCreated, m)
	return nil
}

// get handles GET /api/merchants/{id} with its nested tree preloaded and
// secrets projected for the caller; out-of-scope yields 404 (Req 4.3, 18.2,
// 8.7).
func (h *merchantHandlers) get(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	m, err := h.svc.Get(p, r.PathValue("id"))
	if err != nil {
		return err
	}
	redactMerchantSecrets(p, m)
	middleware.WriteJSON(w, http.StatusOK, m)
	return nil
}

// update handles PUT /api/merchants/{id}, fully replacing the merchant's nested
// tree with the supplied set (Req 8.2). The path id is authoritative; an
// out-of-scope record yields 404 (Req 4.3).
func (h *merchantHandlers) update(w http.ResponseWriter, r *http.Request) error {
	p, err := requirePrincipal(r)
	if err != nil {
		return err
	}
	var m model.Merchant
	if err := decodeJSON(r, &m); err != nil {
		return err
	}
	m.ID = r.PathValue("id")
	if err := h.svc.Update(p, &m); err != nil {
		return err
	}
	redactMerchantSecrets(p, &m)
	middleware.WriteJSON(w, http.StatusOK, m)
	return nil
}

// del handles DELETE /api/merchants/{id}, removing the merchant and cascading
// its nested banks, ATM cards, payment-gateway credentials, and custom fields
// (Req 8.3); an out-of-scope record yields 404 (Req 4.3).
func (h *merchantHandlers) del(w http.ResponseWriter, r *http.Request) error {
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

// merchantSecretsVisible reports whether the secret-bearing fields of a
// merchant's nested data may be serialized to the given principal (Req 8.7).
//
// The tenant scope already prevents any principal from reading a merchant
// outside its own tenant, so this is the second, finer gate within the tenant:
//   - Admin / SuperAdmin (the owning tenant's administrator) always see secrets.
//   - The owning Merchant principal sees the secrets of its own merchant record.
//   - Every other (portal) role has the secrets withheld.
func merchantSecretsVisible(p service.Principal, m *model.Merchant) bool {
	switch p.Role {
	case service.RoleAdmin, service.RoleSuperAdmin:
		return true
	case service.RoleMerchant:
		return p.OwnerID == m.ID
	default:
		return false
	}
}

// redactMerchantSecrets clears the secret-bearing fields (bank login/txn
// passwords and MPIN, ATM card CVV and PIN, gateway-credential password) on the
// merchant and its nested records whenever the principal is not permitted to
// see them (Req 8.7). Those fields carry `omitempty`, so clearing them drops
// them from the JSON response entirely rather than exposing empty placeholders.
//
// The merchant is request-local (freshly loaded or just decoded), so it is
// redacted in place. When secrets are visible, the merchant is left untouched.
func redactMerchantSecrets(p service.Principal, m *model.Merchant) {
	if m == nil || merchantSecretsVisible(p, m) {
		return
	}
	for i := range m.Banks {
		b := &m.Banks[i]
		b.LoginPassword = ""
		b.TxnPassword = ""
		b.Mpin = ""
		for j := range b.AtmCards {
			b.AtmCards[j].Cvv = ""
			b.AtmCards[j].AtmPin = ""
		}
	}
	for i := range m.PaymentGateways {
		m.PaymentGateways[i].Password = ""
	}
}
