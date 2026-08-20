package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/auth"
)

// codeMethodNotAllowed is the machine-readable error code for a 405 response.
// The other routing/error codes reuse apierr's vocabulary; 405 has no service
// equivalent (it is a pure transport condition) so it is defined here.
const codeMethodNotAllowed = "method_not_allowed"

// Deps carries the collaborators the router needs to serve requests. It starts
// with just the Auth_Service (the only dependency the auth endpoints and the
// Auth middleware require). Later tasks extend it with the per-entity CRUD
// services, the Report_Service, and the Lease_Manager as their route groups are
// mounted — see the extension point in routes.
type Deps struct {
	Auth auth.AuthService
	// DB is the GORM database handle the business handlers use to build
	// tenant-scoped repositories for the per-entity CRUD services (tasks 7.x).
	DB *gorm.DB
}

// NewRouter builds the application's HTTP handler: the route table plus the
// global middleware that wraps every request.
//
// Middleware layering (outermost first):
//   - Recover traps any panic and renders a safe 500 (Req 18.4).
//   - structuredRouteErrors converts the ServeMux's default plaintext 404/405
//     responses into the structured APIError body (Req 1.7).
//   - the ServeMux dispatches to a route; per-route middleware (Auth,
//     TenantScope) and the Error adapter are applied when each route is
//     registered.
//
// The returned handler is ready to pass to http.ListenAndServe.
func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, d)
	return middleware.CORS(middleware.Recover(structuredRouteErrors(mux)))
}

// registerRoutes mounts every route on mux. Routes use Go 1.22+ method-aware
// patterns ("POST /api/auth/login"), so an unknown path returns 404 and a known
// path with the wrong method returns 405 — both natively from the ServeMux and
// both rendered as structured bodies by structuredRouteErrors.
func registerRoutes(mux *http.ServeMux, d Deps) {
	ah := &authHandlers{auth: d.Auth}

	// Public: no Session_Token required (Req 6.1).
	mux.Handle("POST /api/auth/login", public(ah.login))

	// Authenticated: any valid Session_Token. These are not tenant-scoped
	// business endpoints, so they need Auth but not TenantScope.
	mux.Handle("POST /api/auth/logout", authenticated(d.Auth, ah.logout)) // Req 6.7
	mux.Handle("GET /api/me", authenticated(d.Auth, ah.me))

	// ---------------------------------------------------------------------
	// Business route groups. These are tenant-scoped and role-guarded: mounted
	// with protected(...) (Auth + TenantScope) so every handler runs with an
	// authenticated, tenant-scoped principal in context, and wrapped with
	// RequireRoles so only Admin/SuperAdmin reach them. Later tasks add their
	// own groups (merchants, transactions, ledgers, reports, leases) here.
	registerCoreEntityRoutes(mux, d)        // gateways, companies, affiliates (7.1), merchants (7.2)
	registerTransactionAndPaymentRoutes(mux, d) // transactions (9.1 wiring), settlements, affiliate/merchant payments (7.3)
	registerLedgerRoutes(mux, d)            // company/affiliate/merchant ledgers (10.1)
	registerReportRoutes(mux, d)            // six report types with date filtering (10.2)
	registerLeaseRoutes(mux, d)             // SuperAdmin-only lease administration (12.3)
	registerAccountRoutes(mux, d)           // SuperAdmin-only account viewing/editing
	registerDashboardRoute(mux, d)          // tenant-scoped summary metrics for every role
	registerPortalRoutes(mux, d)            // read-only owner-scoped reads for portal roles
}

// registerPortalRoutes mounts the /api/portal/* endpoints used by the Company,
// Affiliate, and Merchant portals. The core entity endpoints are Admin/SuperAdmin
// only, so these give portal principals access to the slice of data they own.
//
// Reads are available to all three portal roles and are scoped per role to the
// principal's owned records inside its tenant.
//
// Writes are Company-only and cover the two records a company genuinely
// originates: the transactions its merchants put through, and the settlements it
// receives. Each write handler pins the record's companyId to the principal's
// OwnerID and resolves existing records through the tenant+owner scope, so a
// Company can never touch another company's data. Recording a transaction is
// what drives the commission figures, so the admin's commission and the
// merchant/affiliate split reflect a company-recorded transaction immediately.
// Affiliate and Merchant portals stay read-only.
func registerPortalRoutes(mux *http.ServeMux, d Deps) {
	ph := newPortalHandlers(d)
	portalRoles := RequireRoles(service.RoleCompany, service.RoleAffiliate, service.RoleMerchant)
	companyOnly := RequireRoles(service.RoleCompany)

	mux.Handle("GET /api/portal/merchants", protected(d.Auth, portalRoles(ph.merchants)))
	mux.Handle("GET /api/portal/transactions", protected(d.Auth, portalRoles(ph.transactions)))
	mux.Handle("GET /api/portal/settlements", protected(d.Auth, portalRoles(ph.settlements)))
	mux.Handle("GET /api/portal/affiliate-payments", protected(d.Auth, portalRoles(ph.affiliatePayments)))
	mux.Handle("GET /api/portal/merchant-payments", protected(d.Auth, portalRoles(ph.merchantPayments)))

	// Reference lists the portal forms need (gateway/bank pickers) and the
	// principal's own company (for its gateway commission assignments).
	mux.Handle("GET /api/portal/gateways", protected(d.Auth, portalRoles(ph.gateways)))
	mux.Handle("GET /api/portal/banks", protected(d.Auth, portalRoles(ph.banks)))
	mux.Handle("GET /api/portal/company", protected(d.Auth, companyOnly(ph.company)))

	// Company-only writes.
	mux.Handle("POST /api/portal/transactions", protected(d.Auth, companyOnly(ph.createTransaction)))
	mux.Handle("PUT /api/portal/transactions/{id}", protected(d.Auth, companyOnly(ph.updateTransaction)))
	mux.Handle("DELETE /api/portal/transactions/{id}", protected(d.Auth, companyOnly(ph.deleteTransaction)))

	mux.Handle("POST /api/portal/settlements", protected(d.Auth, companyOnly(ph.createSettlement)))
	mux.Handle("PUT /api/portal/settlements/{id}", protected(d.Auth, companyOnly(ph.updateSettlement)))
	mux.Handle("DELETE /api/portal/settlements/{id}", protected(d.Auth, companyOnly(ph.deleteSettlement)))
}

// registerDashboardRoute mounts the read-only dashboard summary endpoint. It is
// tenant-scoped (protected) and available to every authenticated role; each
// principal only ever receives counts for data within its own tenant/owner
// scope.
func registerDashboardRoute(mux *http.ServeMux, d Deps) {
	dh := newDashboardHandlers(d)
	mux.Handle("GET /api/dashboard", protected(d.Auth, dh.generate))
}

// registerLeaseRoutes mounts the lease administration endpoints (Req 13, 15).
// Every lease route is SuperAdmin-only: it is wrapped with
// RequireRoles(service.RoleSuperAdmin), so any other authenticated role gets a
// 403 (Req 7.4, 15.7) and a request with no valid Session_Token gets a 401 from
// the Auth middleware (Req 7.2). The routes are mounted with protected(...)
// (Auth + TenantScope) for consistency with the other business route groups;
// the Lease_Manager itself is not tenant-scoped, so the SuperAdmin can view and
// manage every lease — including those whose effective status is Expired
// (Req 14.5).
func registerLeaseRoutes(mux *http.ServeMux, d Deps) {
	lh := newLeaseHandlers(d)
	superAdminOnly := RequireRoles(service.RoleSuperAdmin)

	mux.Handle("GET /api/leases", protected(d.Auth, superAdminOnly(lh.list)))
	mux.Handle("POST /api/leases", protected(d.Auth, superAdminOnly(lh.create)))
	mux.Handle("POST /api/leases/{id}/extend", protected(d.Auth, superAdminOnly(lh.extend)))
	mux.Handle("POST /api/leases/{id}/suspend", protected(d.Auth, superAdminOnly(lh.suspend)))
	mux.Handle("POST /api/leases/{id}/reactivate", protected(d.Auth, superAdminOnly(lh.reactivate)))
	mux.Handle("POST /api/leases/{id}/revoke", protected(d.Auth, superAdminOnly(lh.revoke)))
}

// registerReportRoutes mounts the report endpoint (Req 11). It is tenant-scoped
// (protected) so the Report_Service aggregates only over the principal's tenant
// and permitted scope (Req 11.6), and restricted to the Admin and SuperAdmin
// roles. The {type} path segment selects one of the six report types; optional
// start/end query parameters apply inclusive date filtering (Req 11.1, 11.2).
func registerReportRoutes(mux *http.ServeMux, d Deps) {
	rh := newReportHandlers(d)
	adminOnly := RequireRoles(service.RoleAdmin, service.RoleSuperAdmin)
	mux.Handle("GET /api/reports/{type}", protected(d.Auth, adminOnly(rh.generate)))
}

// registerLedgerRoutes mounts the read-only ledger endpoints (Req 10). Each is
// tenant-scoped (protected) so the Report_Service aggregates only over the
// principal's tenant and permitted scope (Req 10.5).
//
// Role guard. Each ledger is restricted to Admin and SuperAdmin plus the one
// portal role it concerns: the company ledger also admits a Company principal,
// the affiliate ledger an Affiliate, the merchant ledger a Merchant. A portal
// principal therefore can reach only the ledger type that applies to it, and
// the tenant + owner scope (repo.ScopeTenant) further restricts the aggregation
// to the records that principal owns — so requesting another owner's ledger
// yields zero balances and never discloses cross-scope data (Req 7.5, 10.5).
func registerLedgerRoutes(mux *http.ServeMux, d Deps) {
	lh := newLedgerHandlers(d)

	companyLedger := RequireRoles(service.RoleAdmin, service.RoleSuperAdmin, service.RoleCompany)
	mux.Handle("GET /api/ledgers/company/{id}", protected(d.Auth, companyLedger(lh.company)))

	affiliateLedger := RequireRoles(service.RoleAdmin, service.RoleSuperAdmin, service.RoleAffiliate)
	mux.Handle("GET /api/ledgers/affiliate/{id}", protected(d.Auth, affiliateLedger(lh.affiliate)))

	merchantLedger := RequireRoles(service.RoleAdmin, service.RoleSuperAdmin, service.RoleMerchant)
	mux.Handle("GET /api/ledgers/merchant/{id}", protected(d.Auth, merchantLedger(lh.merchant)))
}

// registerCoreEntityRoutes mounts the gateway, company, affiliate, and merchant
// CRUD endpoints. Each is tenant-scoped (protected) and restricted to the Admin
// and SuperAdmin roles (Req 8.1); the tenant-scope enforcement in the
// repository layer further limits reads to the principal's tenant (Req 8.5).
// Merchant create/update persist nested data together and delete cascades it
// (Req 8.2, 8.3).
func registerCoreEntityRoutes(mux *http.ServeMux, d Deps) {
	adminOnly := RequireRoles(service.RoleAdmin, service.RoleSuperAdmin)

	gw := newGatewayHandlers(d)
	mux.Handle("GET /api/gateways", protected(d.Auth, adminOnly(gw.list)))
	mux.Handle("POST /api/gateways", protected(d.Auth, adminOnly(gw.create)))
	mux.Handle("GET /api/gateways/{id}", protected(d.Auth, adminOnly(gw.get)))
	mux.Handle("PUT /api/gateways/{id}", protected(d.Auth, adminOnly(gw.update)))
	mux.Handle("DELETE /api/gateways/{id}", protected(d.Auth, adminOnly(gw.del)))

	co := newCompanyHandlers(d)
	mux.Handle("GET /api/companies", protected(d.Auth, adminOnly(co.list)))
	mux.Handle("POST /api/companies", protected(d.Auth, adminOnly(co.create)))
	mux.Handle("GET /api/companies/{id}", protected(d.Auth, adminOnly(co.get)))
	mux.Handle("PUT /api/companies/{id}", protected(d.Auth, adminOnly(co.update)))
	mux.Handle("DELETE /api/companies/{id}", protected(d.Auth, adminOnly(co.del)))

	af := newAffiliateHandlers(d)
	mux.Handle("GET /api/affiliates", protected(d.Auth, adminOnly(af.list)))
	mux.Handle("POST /api/affiliates", protected(d.Auth, adminOnly(af.create)))
	mux.Handle("GET /api/affiliates/{id}", protected(d.Auth, adminOnly(af.get)))
	mux.Handle("PUT /api/affiliates/{id}", protected(d.Auth, adminOnly(af.update)))
	mux.Handle("DELETE /api/affiliates/{id}", protected(d.Auth, adminOnly(af.del)))

	me := newMerchantHandlers(d)
	mux.Handle("GET /api/merchants", protected(d.Auth, adminOnly(me.list)))
	mux.Handle("POST /api/merchants", protected(d.Auth, adminOnly(me.create)))
	mux.Handle("GET /api/merchants/{id}", protected(d.Auth, adminOnly(me.get)))
	mux.Handle("PUT /api/merchants/{id}", protected(d.Auth, adminOnly(me.update)))
	mux.Handle("DELETE /api/merchants/{id}", protected(d.Auth, adminOnly(me.del)))

	bk := newBankHandlers(d)
	mux.Handle("GET /api/banks", protected(d.Auth, adminOnly(bk.list)))
	mux.Handle("POST /api/banks", protected(d.Auth, adminOnly(bk.create)))
	mux.Handle("GET /api/banks/{id}", protected(d.Auth, adminOnly(bk.get)))
	mux.Handle("PUT /api/banks/{id}", protected(d.Auth, adminOnly(bk.update)))
	mux.Handle("DELETE /api/banks/{id}", protected(d.Auth, adminOnly(bk.del)))
}

// registerTransactionAndPaymentRoutes mounts the transaction, settlement,
// affiliate-payment, and merchant-payment CRUD endpoints (task 7.3, plus the
// task 9.1 transaction breakdown wiring). Each group is tenant-scoped
// (protected) and restricted to the Admin and SuperAdmin roles (Req 8.1); the
// tenant-scope enforcement in the repository layer further limits reads to the
// principal's tenant (Req 8.5).
//
// Transaction reads (list and get) embed the commission breakdown computed by
// the Commission_Engine from the transaction's related Company/Merchant/
// Affiliate, so the engine ported in task 9.1 is no longer orphaned (Req 9).
func registerTransactionAndPaymentRoutes(mux *http.ServeMux, d Deps) {
	adminOnly := RequireRoles(service.RoleAdmin, service.RoleSuperAdmin)

	tx := newTransactionHandlers(d)
	mux.Handle("GET /api/transactions", protected(d.Auth, adminOnly(tx.list)))
	mux.Handle("POST /api/transactions", protected(d.Auth, adminOnly(tx.create)))
	mux.Handle("GET /api/transactions/{id}", protected(d.Auth, adminOnly(tx.get)))
	mux.Handle("PUT /api/transactions/{id}", protected(d.Auth, adminOnly(tx.update)))
	mux.Handle("DELETE /api/transactions/{id}", protected(d.Auth, adminOnly(tx.del)))

	st := newSettlementHandlers(d)
	mux.Handle("GET /api/settlements", protected(d.Auth, adminOnly(st.list)))
	mux.Handle("POST /api/settlements", protected(d.Auth, adminOnly(st.create)))
	mux.Handle("GET /api/settlements/{id}", protected(d.Auth, adminOnly(st.get)))
	mux.Handle("PUT /api/settlements/{id}", protected(d.Auth, adminOnly(st.update)))
	mux.Handle("DELETE /api/settlements/{id}", protected(d.Auth, adminOnly(st.del)))

	ap := newAffiliatePaymentHandlers(d)
	mux.Handle("GET /api/affiliate-payments", protected(d.Auth, adminOnly(ap.list)))
	mux.Handle("POST /api/affiliate-payments", protected(d.Auth, adminOnly(ap.create)))
	mux.Handle("GET /api/affiliate-payments/{id}", protected(d.Auth, adminOnly(ap.get)))
	mux.Handle("PUT /api/affiliate-payments/{id}", protected(d.Auth, adminOnly(ap.update)))
	mux.Handle("DELETE /api/affiliate-payments/{id}", protected(d.Auth, adminOnly(ap.del)))

	mp := newMerchantPaymentHandlers(d)
	mux.Handle("GET /api/merchant-payments", protected(d.Auth, adminOnly(mp.list)))
	mux.Handle("POST /api/merchant-payments", protected(d.Auth, adminOnly(mp.create)))
	mux.Handle("GET /api/merchant-payments/{id}", protected(d.Auth, adminOnly(mp.get)))
	mux.Handle("PUT /api/merchant-payments/{id}", protected(d.Auth, adminOnly(mp.update)))
	mux.Handle("DELETE /api/merchant-payments/{id}", protected(d.Auth, adminOnly(mp.del)))

	cp := newCompanyPaymentHandlers(d)
	mux.Handle("GET /api/company-payments", protected(d.Auth, adminOnly(cp.list)))
	mux.Handle("POST /api/company-payments", protected(d.Auth, adminOnly(cp.create)))
	mux.Handle("GET /api/company-payments/{id}", protected(d.Auth, adminOnly(cp.get)))
	mux.Handle("DELETE /api/company-payments/{id}", protected(d.Auth, adminOnly(cp.del)))
}

// public adapts an error-returning handler into an http.Handler with no
// authentication. Errors are rendered by the Error middleware.
func public(h middleware.HandlerFunc) http.Handler {
	return middleware.Error(h)
}

// authenticated wraps a handler so it runs only behind a valid Session_Token.
// The Auth middleware validates the token and places the principal in context;
// the Error adapter renders the handler's typed errors. Used for endpoints that
// require authentication but are not tenant-scoped business data (logout, me).
func authenticated(svc auth.AuthService, h middleware.HandlerFunc) http.Handler {
	return middleware.Auth(svc)(middleware.Error(h))
}

// protected wraps a handler with Auth then TenantScope, so the handler runs
// only for an authenticated principal and the tenant-scoped repository layer
// always has a scope to apply (Req 4.5, 7.6). This is the chain the business
// route groups in later tasks use; combine it with RequireRoles to enforce
// role access on top of authentication.
//
//nolint:unused // reusable extension point consumed by later route groups (tasks 7.x, 10.x, 12.x).
func protected(svc auth.AuthService, h middleware.HandlerFunc) http.Handler {
	return middleware.Auth(svc)(middleware.TenantScope(middleware.Error(h)))
}

// decodeJSON decodes the JSON request body into dst. A body that cannot be
// parsed is a client error, so it is reported as a 400 validation error (Req
// 18.3) rather than surfacing the raw decode error.
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apierr.Validation("request body is required", map[string]string{"body": "must be a JSON object"})
	}
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apierr.Validation("request body is not valid JSON", map[string]string{"body": "must be a JSON object"})
	}
	return nil
}

// structuredRouteErrors converts the ServeMux's built-in 404 (unknown route)
// and 405 (unsupported method) responses — which it writes as plaintext — into
// the structured APIError JSON body required by Req 1.7, while preserving the
// status code and any Allow header the mux set.
//
// It does not register a catch-all route, so the ServeMux's native distinction
// between "no such path" (404) and "path exists, wrong method" (405) is kept
// intact. Responses that handlers write themselves (which set a JSON
// Content-Type via WriteJSON) pass through untouched.
func structuredRouteErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&routeErrorInterceptor{ResponseWriter: w}, r)
	})
}

// routeErrorInterceptor wraps a ResponseWriter to rewrite the ServeMux default
// 404/405 plaintext error into a structured JSON body. It keys off the
// Content-Type: the mux default path sets text/plain, whereas every application
// response goes through WriteJSON and sets application/json, so only the mux's
// own error output is rewritten.
type routeErrorInterceptor struct {
	http.ResponseWriter
	swallow bool
}

func (w *routeErrorInterceptor) WriteHeader(status int) {
	if status == http.StatusNotFound || status == http.StatusMethodNotAllowed {
		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
			// ServeMux default (plaintext) error path: replace the body with a
			// structured APIError. The status and any Allow header are kept.
			w.swallow = true
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.ResponseWriter.WriteHeader(status)
			_ = json.NewEncoder(w.ResponseWriter).Encode(routeErrorBody(status))
			return
		}
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *routeErrorInterceptor) Write(b []byte) (int, error) {
	if w.swallow {
		// Discard the mux's plaintext body; the structured body was already
		// written in WriteHeader. Report success so the mux does not error.
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

// routeErrorBody is the structured error body for an unmatched route (404) or
// an unsupported method (405).
func routeErrorBody(status int) apierr.APIError {
	if status == http.StatusMethodNotAllowed {
		return apierr.APIError{
			Code:    codeMethodNotAllowed,
			Message: "the request method is not allowed for this resource",
		}
	}
	return apierr.APIError{
		Code:    apierr.CodeNotFound,
		Message: "the requested resource was not found",
	}
}
