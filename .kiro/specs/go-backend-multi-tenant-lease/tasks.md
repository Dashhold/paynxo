# Implementation Plan: Go Backend with Multi-Tenant Leasing

## Overview

This plan converts the design into incremental Go backend and React frontend coding steps. It starts with project scaffolding and configuration, builds the GORM data model and migrations, then layers authentication, tenant-scope enforcement, CRUD services/handlers, the commission engine, ledger/report services, and the lease manager. Each step wires its output into the running server (router/middleware) so there is no orphaned code. The 21 correctness properties from the design are implemented as `pgregory.net/rapid` property-based tests (minimum 100 iterations, tagged `// Feature: go-backend-multi-tenant-lease, Property {number}: {property_text}`) placed next to the code they validate. Finally, the React data layer is migrated from localStorage to the API while preserving every screen.

Implementation language: **Go** (backend) and **JavaScript/React + Vite** (frontend), as specified by the design.

## Tasks

- [x] 1. Scaffold the Go backend project and configuration
  - [x] 1.1 Initialize the Go module and package layout
    - Create `backend/go.mod` (module path, Go version) and the directory skeleton: `cmd/server`, `internal/config`, `internal/model`, `internal/repo`, `internal/service`, `internal/middleware`, `internal/api`
    - Add dependencies: GORM + Postgres driver, JWT, bcrypt (`golang.org/x/crypto/bcrypt`), `pgregory.net/rapid`, a CSV writer (stdlib `encoding/csv`) and a PDF library
    - Add a placeholder `cmd/server/main.go` that compiles
    - _Requirements: 1.1, 1.3_

  - [x] 1.2 Implement config loading and validation
    - Implement `internal/config.Config`, `Load()` (read `DB_HOST`/`DB_PORT`/`DB_NAME`/`DB_USER`/`DB_PASSWORD`/`PORT`/`TOKEN_SECRET`/`TOKEN_TTL` from env with documented defaults), and `Redacted()`
    - `Load()` returns an error naming the first missing required value (`DB_PASSWORD`, `TOKEN_SECRET`); `Redacted()` masks secrets as `***`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [ ]* 1.3 Write property test for configuration validation and secret redaction
    - **Property 21: Configuration validation and secret redaction**
    - **Validates: Requirements 2.4, 2.5**

- [x] 2. Define GORM models and run migrations
  - [x] 2.1 Implement the data model structs
    - Define `TenantBase`, `Tenant`, `Account`, `Gateway`, `Company`, `CompanyGateway`, `Affiliate`, `Merchant`, `MerchantBank`, `AtmCard`, `MerchantGatewayCredential`, `CustomField`, `Transaction`, `Settlement`, `AffiliatePayment`, `MerchantPayment`, `Lease`, `RevokedToken` in `internal/model`
    - Preserve every field from `seed.js`/`calc.js` shapes (nested banks, ATM cards, credentials, custom fields, `AffiliateID *string`, `MerchantRef`, etc.)
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

  - [x] 2.2 Implement database connection and Migration_Service.AutoMigrate
    - Implement `internal/service/migration` with a DB-connect helper and `AutoMigrate(models...)` creating all tables, columns, indexes, and foreign keys
    - Wire connect + migrate into `cmd/server/main.go` startup; on connection or migration failure log a descriptive error and `exit(1)` before serving
    - _Requirements: 1.2, 1.4, 1.5, 1.6, 5.1_

- [x] 3. Implement the tenant-scoped repository layer
  - [x] 3.1 Implement the Principal type and tenant/owner GORM scope
    - Define `service.Principal` and `repo.ScopeTenant(p)` applying `tenant_id` plus owner filters for Company/Affiliate/Merchant
    - Implement generic tenant-scoped repository helpers (list/get-by-id/create/update/delete) that always apply the scope and set `tenant_id` from the principal on create (ignoring client-supplied tenant); get-by-id outside scope returns a typed not-found error
    - _Requirements: 4.1, 4.2, 4.3, 4.5, 7.5_

  - [ ]* 3.2 Write property test for tenant read isolation
    - **Property 4: Tenant read isolation**
    - **Validates: Requirements 4.1, 4.3, 4.4, 4.5, 8.5, 10.5, 11.6, 12.3, 18.2**

  - [ ]* 3.3 Write property test for tenant assignment on create
    - **Property 5: Tenant assignment on create**
    - **Validates: Requirements 4.2**

  - [ ]* 3.4 Write property test for portal ownership scoping
    - **Property 6: Portal ownership scoping**
    - **Validates: Requirements 7.5, 8.5**

- [x] 4. Implement the API error model and middleware
  - [x] 4.1 Implement the structured error model and typed service errors
    - Define `APIError{code,message,fields}` JSON model and typed errors (`ErrNotFound`, `ErrValidation`, `ErrConflict`, `ErrForbidden`, `ErrUnauthenticated`, `ErrLeaseInactive`)
    - Implement `Error` middleware mapping typed errors to HTTP status (400/401/403/404/409/500) and `Recover` middleware that traps panics and returns 500 without stack traces or secrets
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5_

  - [ ]* 4.2 Write property test for validation errors returning 400 with the offending field
    - **Property 18: Validation errors return 400 with the offending field**
    - **Validates: Requirements 8.4, 18.3**

  - [ ]* 4.3 Write property test for structured error body and safe 500s
    - **Property 19: Structured error body and safe 500s**
    - **Validates: Requirements 18.1, 18.4**

- [x] 5. Implement authentication and token middleware
  - [x] 5.1 Implement Auth_Service login, authenticate, and logout
    - Implement bcrypt password verification, JWT issuance (`account_id`, `role`, `tenant_id`, owner fields, `jti`, `exp`) signed with `TOKEN_SECRET`, `Authenticate(token)` validating signature/expiry and checking the `revoked_tokens` set, and `Logout` inserting the `jti`
    - Return generic 401 `invalid_credentials` on bad login (no field disclosure); 401 `unauthenticated` on missing/invalid/expired/revoked tokens
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

  - [x] 5.2 Implement Auth and TenantScope middleware
    - Implement `Auth` middleware that authenticates the token, loads the principal, and (for leased Admins) resolves lease status — placing the principal in request context
    - Implement `TenantScope` middleware that derives `tenant_id` and owner scope into context for repositories; reject missing token with 401, expired/suspended/revoked lease with 403
    - _Requirements: 7.1, 7.2, 7.6, 14.3, 15.4, 15.6_

  - [ ]* 5.3 Write property test for passwords stored only as one-way hashes
    - **Property 9: Passwords are stored only as one-way hashes**
    - **Validates: Requirements 5.5, 6.3, 6.4, 13.2**

  - [ ]* 5.4 Write property test for token issue/authenticate round-trip
    - **Property 10: Token issue/authenticate round-trip**
    - **Validates: Requirements 6.1, 6.5**

  - [ ]* 5.5 Write property test for token rejection when invalid, expired, or revoked
    - **Property 11: Tokens are rejected when invalid, expired, or revoked**
    - **Validates: Requirements 6.6, 6.7**

  - [ ]* 5.6 Write property test for indistinguishable invalid-credential 401
    - **Property 12: Invalid credentials yield an indistinguishable 401**
    - **Validates: Requirements 6.2**

- [x] 6. Wire the router and auth endpoints
  - [x] 6.1 Implement the HTTP router with auth endpoints and role guards
    - Build the router in `internal/api`, mount `Recover`/`Error`/`Auth`/`TenantScope` middleware, add `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/me`
    - Implement a role-authorization helper for handlers; unknown route → 404 and unsupported method → 405 with structured bodies
    - _Requirements: 1.7, 6.1, 6.7, 7.2, 7.3, 7.6_

  - [ ]* 6.2 Write property test for role-based authorization
    - **Property 13: Role-based authorization**
    - **Validates: Requirements 7.3, 7.4, 15.7**

  - [ ]* 6.3 Write unit tests for routing error responses
    - Unknown route → 404, unsupported method → 405 with structured error body
    - _Requirements: 1.7_

- [x] 7. Implement CRUD services and handlers for core entities
  - [x] 7.1 Implement gateway, company (with gateway assignments), and affiliate CRUD
    - Implement tenant-scoped services + handlers for `/api/gateways`, `/api/companies` (persist `CompanyGateway` assignments), `/api/affiliates`; validate required fields returning 400 with the invalid field
    - Wire routes into the router
    - _Requirements: 8.1, 8.4, 8.5, 8.6_

  - [x] 7.2 Implement merchant CRUD with nested data persistence and cascade delete
    - Implement `/api/merchants` create/update persisting nested banks, ATM cards, payment-gateway credentials, and custom fields in one operation; delete cascades nested records; omit secret-bearing fields from DTOs for non-owning principals
    - Wire routes into the router
    - _Requirements: 8.1, 8.2, 8.3, 8.7_

  - [x] 7.3 Implement settlement, affiliate-payment, and merchant-payment CRUD
    - Implement tenant-scoped services + handlers for `/api/settlements`, `/api/affiliate-payments`, `/api/merchant-payments` and wire routes
    - _Requirements: 8.1, 8.4, 8.5_

  - [ ]* 7.4 Write property test for persistence round-trip preserving fields and nested data
    - **Property 7: Persistence round-trip preserves all fields and nested data**
    - **Validates: Requirements 3.3, 3.4, 3.5, 8.1, 8.2**

  - [ ]* 7.5 Write property test for cascade delete of nested merchant data
    - **Property 8: Cascade delete removes nested merchant data**
    - **Validates: Requirements 8.3**

  - [ ]* 7.6 Write unit tests for out-of-scope reads and secret-field omission
    - Out-of-scope-but-in-tenant read returns empty set and logs the attempt; secret fields omitted for non-owning portal roles
    - _Requirements: 8.6, 8.7_

- [x] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement the Commission_Engine
  - [x] 9.1 Implement Commission_Engine.Calc as a field-for-field port of calc.js
    - Implement `internal/service/commission.Calc(txn, ctx)` producing the full `Breakdown` (admin commission, beneficiary selection, base selection, company net income, admin net income, charge-bearer deductions) per the calc.js algorithm
    - Expose computed breakdown on transaction reads via the transactions handler (`GET /api/transactions`, `/{id}`)
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7_

  - [ ]* 9.2 Write property test for commission engine equivalence to calc.js
    - **Property 1: Commission engine equivalence to calc.js**
    - Use a faithful Go transliteration / golden oracle of `calc.js` as the equivalence reference
    - **Validates: Requirements 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7**

- [x] 10. Implement the Report_Service (ledgers, reports, exports)
  - [x] 10.1 Implement tenant-scoped ledger aggregation
    - Implement `CompanyLedger`, `AffiliateLedger`, `MerchantLedger` over scoped records (direct merchants only); wire `GET /api/ledgers/company/{id}`, `/affiliate/{id}`, `/merchant/{id}`
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

  - [x] 10.2 Implement the six report types with date filtering
    - Implement `Generate(req, scope)` for company/merchant/affiliate/gateway/settlement/outstanding reports with inclusive `[start,end]` date filtering; wire `GET /api/reports/{type}?start&end&format`
    - _Requirements: 11.1, 11.2, 11.6_

  - [x] 10.3 Implement CSV and PDF exports with independent format handling
    - Implement `ExportCSV` and `ExportPDF`; the report endpoint produces each requested format independently so a failure in one format does not block another
    - _Requirements: 11.3, 11.4, 11.5_

  - [ ]* 10.4 Write property test for ledger aggregation correctness
    - **Property 2: Ledger aggregation correctness**
    - **Validates: Requirements 10.1, 10.2, 10.3, 10.4**

  - [ ]* 10.5 Write property test for report date-range filtering
    - **Property 3: Report date-range filtering**
    - **Validates: Requirements 11.2**

  - [ ]* 10.6 Write unit tests for report types and export formats
    - Each of the six report types returns expected columns; CSV header/rows correct; PDF produces valid bytes; one failing format does not block another
    - _Requirements: 11.1, 11.3, 11.4, 11.5_

- [x] 11. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 12. Implement the Lease_Manager and lease endpoints
  - [x] 12.1 Implement the lease status state machine
    - Implement `LeaseStatus`, `EffectiveStatus(l, now)` with precedence Revoked → Suspended → Expired(now>expiry) → Active
    - _Requirements: 13.3, 14.1, 15.2, 15.3, 15.5, 15.6_

  - [x] 12.2 Implement lease create, list, extend, suspend, reactivate, revoke
    - Implement `Create` (validate fields, `expiry>start` else 400, duplicate userID 409, validation-precedence 400) in one transaction creating a new `Tenant`, bcrypt-hashed Admin `Account`, and `Lease` (status Active); implement `List`, `Extend`, `Suspend`, `Reactivate`, `Revoke`
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7, 14.4, 15.1, 15.2, 15.3, 15.5, 15.6_

  - [x] 12.3 Wire SuperAdmin-only lease endpoints and enforce access denial
    - Mount `GET/POST /api/leases`, `POST /api/leases/{id}/extend|suspend|reactivate|revoke` guarded to SuperAdmin (403 otherwise); ensure Auth middleware denies expired/suspended/revoked leased Admins at login and on token use; SuperAdmin can still view/manage expired leases
    - _Requirements: 7.4, 14.2, 14.3, 14.5, 15.4, 15.6, 15.7_

  - [ ]* 12.4 Write property test for the lease status state machine
    - **Property 14: Lease status state machine**
    - **Validates: Requirements 13.3, 14.1, 15.2, 15.3, 15.5, 15.6**

  - [ ]* 12.5 Write property test for non-active leases denying access
    - **Property 15: Non-active leases deny access**
    - **Validates: Requirements 14.2, 14.3, 15.4, 15.6**

  - [ ]* 12.6 Write property test for lease creation validation and conflict precedence
    - **Property 16: Lease creation validation and conflict precedence**
    - **Validates: Requirements 13.4, 13.5, 13.6**

  - [ ]* 12.7 Write property test for new leased tenant being empty and isolated
    - **Property 17: New leased tenant is empty and isolated**
    - **Validates: Requirements 13.1, 13.7, 14.4**

  - [ ]* 12.8 Write unit test for SuperAdmin viewing/managing an expired lease
    - SuperAdmin can list and manage a lease whose effective status is Expired
    - _Requirements: 14.5_

- [x] 13. Implement migration seeding
  - [x] 13.1 Implement SeedIfEmpty with idempotent demo data
    - Implement `SeedIfEmpty()`: sentinel detection (demo tenant), create SuperAdmin (own tenant) and demo Admin (separate tenant) with bcrypt-hashed documented credentials, map `seed.js` data into the demo Admin tenant; wire into startup after AutoMigrate
    - _Requirements: 5.2, 5.3, 5.4, 5.5, 12.1, 12.2, 12.3, 12.4_

  - [ ]* 13.2 Write property test for idempotent seeding
    - **Property 20: Seeding is idempotent**
    - **Validates: Requirements 5.4**

  - [ ]* 13.3 Write unit tests for seeding content
    - Demo records present under demo tenant; SuperAdmin + demo Admin created in separate tenants
    - _Requirements: 5.2, 5.3, 12.2, 12.4_

- [x] 14. Backend integration tests
  - [ ]* 14.1 Write startup and integration tests against PostgreSQL
    - Server starts/migrates/serves over HTTP against a disposable Postgres; bad DB connection → exit; migration failure → exit
    - _Requirements: 1.1, 1.2, 1.4, 1.5, 1.6, 5.1_

  - [ ]* 14.2 Write end-to-end auth and authorization-matrix integration tests
    - login → authenticated request → logout → rejected; endpoint authorization matrix across all roles
    - _Requirements: 6.1, 6.5, 6.6, 6.7, 7.1, 7.3, 7.6_

- [x] 15. Checkpoint - Ensure all backend tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 16. Migrate the React frontend data layer to the API
  - [x] 16.1 Implement apiClient with token handling and 401 redirect
    - Create `frontend/src/data/apiClient.js` with `setToken`/`clearToken`, an `api(method, path, body)` wrapper attaching the bearer token and `ApiError`, and a 401 handler that clears the token and redirects to login
    - _Requirements: 17.2, 17.6_

  - [x] 16.2 Rewrite store.jsx to back useStore() with the API
    - Rewrite `frontend/src/data/store.jsx` to keep the `useStore()` contract (`db`, `auth`, `add`, `update`, `remove`, `login`, `logout`) but call `api()` instead of localStorage; `login` stores the token, `logout` clears it; commission/ledger/report values come from the server
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5_

  - [x] 16.3 Add SuperAdmin Leases navigation and lease management screen
    - Add a `SuperAdmin` entry in `Layout.jsx` `NAV_BY_ROLE` equal to Admin nav plus `{ to: '/leases', label: 'Leases' }` immediately below Reports; other roles never see it; add the `/leases` route and a lease management screen (list/create/extend/suspend/reactivate/revoke) consuming the lease endpoints
    - _Requirements: 16.1, 16.2, 16.3, 16.4_

  - [ ]* 16.4 Write frontend tests for navigation and data-layer behavior
    - SuperAdmin nav shows Leases below Reports and other roles do not; requests carry the token; login stores and logout clears it; 401 redirects to login
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 17.2, 17.5, 17.6_

  - [ ]* 16.5 Write frontend regression smoke tests for existing screens
    - Existing Admin/Company/Affiliate/Merchant screens render against the API and display server-computed commission/ledger/report values
    - _Requirements: 17.1, 17.3, 17.4_

- [x] 17. Final checkpoint - Ensure all tests pass
  - Ensure all backend and frontend tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional test sub-tasks and can be skipped for a faster MVP; core implementation sub-tasks are never optional.
- Each task references specific requirement clauses (and design properties where applicable) for traceability.
- Property-based tests use `pgregory.net/rapid` with a minimum of 100 iterations and the tag `// Feature: go-backend-multi-tenant-lease, Property {number}: {property_text}`.
- Property tests are placed next to the implementation they validate so errors are caught early; each of the 21 properties is implemented by exactly one property-based test.
- Checkpoints provide incremental validation at natural boundaries (repositories/CRUD, computation, leasing, frontend).

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "2.1"] },
    { "id": 2, "tasks": ["1.3", "2.2", "4.1"] },
    { "id": 3, "tasks": ["3.1", "4.2", "4.3"] },
    { "id": 4, "tasks": ["3.2", "3.3", "3.4", "5.1"] },
    { "id": 5, "tasks": ["5.2", "5.3", "5.4", "5.5", "5.6"] },
    { "id": 6, "tasks": ["6.1"] },
    { "id": 7, "tasks": ["6.2", "6.3", "7.1", "7.3", "9.1", "10.1", "12.1"] },
    { "id": 8, "tasks": ["7.2", "9.2", "10.2", "12.2"] },
    { "id": 9, "tasks": ["7.4", "7.5", "7.6", "10.3", "10.4", "10.5", "12.3", "13.1"] },
    { "id": 10, "tasks": ["10.6", "12.4", "12.5", "12.6", "12.7", "12.8", "13.2", "13.3"] },
    { "id": 11, "tasks": ["14.1", "14.2"] },
    { "id": 12, "tasks": ["16.1"] },
    { "id": 13, "tasks": ["16.2", "16.3"] },
    { "id": 14, "tasks": ["16.4", "16.5"] }
  ]
}
```
