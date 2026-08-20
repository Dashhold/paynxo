# Requirements Document

## Introduction

This feature introduces a real backend for the Payment Gateway Commission & Settlement Management System, replacing the current frontend-only localStorage persistence with a Go REST API backed by PostgreSQL (using GORM for models and migrations). The backend takes over all data storage, the commission calculation engine, and ledger/report aggregation that currently live in the browser. The React frontend is migrated to consume the API while preserving every existing screen and behavior.

The feature also adds multi-tenant leasing. A new SuperAdmin role operates its own business exactly like a normal Admin, and can additionally lease isolated instances of the platform to new Admin accounts. Each leased Admin runs an independent, fully isolated business (its own companies, affiliates, merchants, gateways, transactions, settlements, ledgers, and reports) for a defined tenure, after which access is restricted.

This system is being built for a client demo while moving toward a real production backend. Several existing demo behaviors (notably plaintext credentials and unauthenticated data access) are explicitly called out as security requirements to be corrected.

## Glossary

- **API_Server**: The Go REST API service that handles all HTTP requests, persistence, business logic, and authorization for the system.
- **Auth_Service**: The component of the API_Server responsible for authenticating credentials, issuing tokens, and enforcing role-based and tenant-based authorization.
- **Commission_Engine**: The server-side component that computes per-transaction commission breakdowns, equivalent to the current frontend `calc.js` logic.
- **Report_Service**: The server-side component that aggregates ledger balances and report data and produces CSV and PDF exports.
- **Migration_Service**: The component that creates and evolves the PostgreSQL schema via GORM and seeds initial demo data.
- **Web_Client**: The React (Vite) single-page application that consumes the API_Server.
- **SuperAdmin**: A role that sits above Admin. A SuperAdmin operates its own tenant business with full Admin functionality and additionally manages leases for other Admins.
- **Admin**: A role that operates a single independent business (its own companies, affiliates, merchants, gateways, transactions, settlements, ledgers, and reports). Each Admin corresponds to exactly one Tenant.
- **Company**, **Affiliate**, **Merchant**: Existing portal roles that log into their own scoped portals, unchanged in behavior.
- **Tenant**: An isolation boundary that owns a set of business entities. Every Admin (including the SuperAdmin's own business) maps to exactly one Tenant. Data belonging to one Tenant is not visible to any other Tenant.
- **Tenant_Scope**: The enforcement mechanism that constrains every data read and write to the Tenant associated with the authenticated principal.
- **Lease**: A grant from a SuperAdmin to an Admin that authorizes that Admin to operate a Tenant for a defined period. A Lease has a start date, an expiry date (tenure), and a status.
- **Lease_Manager**: The server-side component that creates leases, enforces tenure, and supports extend, suspend, and revoke operations.
- **Tenure**: The active period of a Lease, defined by a start date and an end/expiry date.
- **Lease_Status**: The state of a Lease, one of Active, Expired, Suspended, or Revoked.
- **Session_Token**: A bearer token issued on successful login that identifies the principal, role, and Tenant for subsequent requests.

## Requirements

### Requirement 1: Backend service and database foundation

**User Story:** As a system operator, I want a Go REST API backed by PostgreSQL with GORM, so that application data is persisted server-side instead of in browser localStorage.

#### Acceptance Criteria

1. THE API_Server SHALL expose a REST API over HTTP for all application data and operations.
2. THE API_Server SHALL persist all application data in a PostgreSQL database.
3. THE API_Server SHALL define all database models and relationships using GORM.
4. WHEN the API_Server starts, THE Migration_Service SHALL apply any pending schema migrations before the API_Server begins accepting requests.
5. IF a database connection cannot be established at startup, THEN THE API_Server SHALL log a descriptive error and exit without accepting requests.
6. IF a schema migration fails during startup, THEN THE API_Server SHALL fail startup and exit without accepting requests.
7. WHEN a client sends a request with an unsupported route or method, THE API_Server SHALL respond with HTTP status 404 or 405 and a structured error body.

### Requirement 2: Configuration and environment

**User Story:** As a system operator, I want database and runtime settings provided through configuration, so that I can run the backend across environments without code changes.

#### Acceptance Criteria

1. THE API_Server SHALL read the PostgreSQL connection settings (host, port, database name, user, password) from environment variables or a configuration file.
2. THE API_Server SHALL read the HTTP listen port from configuration with a documented default value.
3. THE API_Server SHALL read the token signing secret and token lifetime from configuration.
4. IF a required configuration value is missing at startup, THEN THE API_Server SHALL attempt to log which value is missing and SHALL exit without accepting requests even if the logging itself fails.
5. THE API_Server SHALL NOT log secret configuration values (database password, token signing secret) in plaintext.

### Requirement 3: Data model and multi-tenancy

**User Story:** As a SuperAdmin, I want every business entity scoped to a tenant, so that each leased Admin's business data is isolated from every other tenant.

#### Acceptance Criteria

1. THE API_Server SHALL model all existing entities: gateways, companies, affiliates, merchants, merchant banks, merchant ATM cards, merchant payment-gateway credentials, transactions, settlements, affiliate payments, and merchant payments.
2. THE API_Server SHALL associate every business entity record with exactly one Tenant via a tenant identifier.
3. THE API_Server SHALL model nested merchant data (banks, ATM cards, custom fields, and payment-gateway credentials) as related records owned by their parent merchant.
4. THE API_Server SHALL preserve all existing fields for each entity as defined by the current frontend data model.
5. WHERE a merchant is assigned to an affiliate, THE API_Server SHALL store the affiliate association on the merchant record.
6. THE API_Server SHALL associate each Admin account with exactly one Tenant.

### Requirement 4: Tenant isolation enforcement

**User Story:** As a leased Admin, I want to see and modify only my own business data, so that no other tenant can access my records and I cannot access theirs.

#### Acceptance Criteria

1. WHEN an authenticated principal requests business data, THE Tenant_Scope SHALL restrict the results to records belonging to the principal's Tenant.
2. WHEN an authenticated principal creates a business record, THE Tenant_Scope SHALL assign the record to the principal's Tenant.
3. IF a principal requests or modifies a record belonging to a different Tenant, THEN THE API_Server SHALL respond with HTTP status 404 and SHALL NOT disclose the record.
4. WHEN the SuperAdmin operates its own business, THE Tenant_Scope SHALL restrict those operations to the SuperAdmin's own Tenant in the same manner as any Admin.
5. THE Tenant_Scope SHALL apply to every endpoint that reads or writes business data.

### Requirement 5: Schema migration and demo seeding

**User Story:** As a developer, I want the database schema created and seeded with demo data, so that the client demo starts with representative content.

#### Acceptance Criteria

1. WHEN migrations run against an empty database, THE Migration_Service SHALL create all tables, columns, indexes, and foreign keys required by the data model.
2. WHEN seeding runs against an empty database, THE Migration_Service SHALL insert the existing demo records (gateways, companies, affiliates, merchants with nested data, transactions, settlements, affiliate payments, merchant payments) assigned to a demo Tenant.
3. WHEN seeding runs, THE Migration_Service SHALL create the SuperAdmin demo account and the existing demo Admin account.
4. IF seeding runs against a database that already contains seeded records, THEN THE Migration_Service SHALL leave existing records unchanged.
5. WHEN seeding creates accounts, THE Migration_Service SHALL store each account password using a one-way hash rather than plaintext.

### Requirement 6: Authentication and password security

**User Story:** As any user, I want to log in with my user id and password and receive a token, so that my subsequent requests are authenticated securely.

#### Acceptance Criteria

1. WHEN a client submits a valid user id and password to the login endpoint, THE Auth_Service SHALL issue a Session_Token identifying the principal, role, and Tenant.
2. IF a client submits credentials that do not match any account, THEN THE Auth_Service SHALL respond with HTTP status 401 and SHALL NOT indicate whether the user id or the password was incorrect.
3. THE Auth_Service SHALL verify submitted passwords against stored one-way password hashes.
4. THE Auth_Service SHALL store all account passwords as one-way hashes and SHALL NOT store passwords in plaintext.
5. WHEN a request includes a valid Session_Token, THE Auth_Service SHALL authenticate the request as the token's principal.
6. IF a request includes an expired or invalid Session_Token, THEN THE Auth_Service SHALL respond with HTTP status 401.
7. WHEN a client requests logout, THE Auth_Service SHALL invalidate the client's active Session_Token.

### Requirement 7: Role-based authorization

**User Story:** As a system operator, I want each role limited to its permitted operations, so that users cannot access functionality outside their role.

#### Acceptance Criteria

1. THE Auth_Service SHALL recognize the roles SuperAdmin, Admin, Company, Affiliate, and Merchant.
2. IF a request requires no valid Session_Token but none is present, THEN THE API_Server SHALL respond with HTTP status 401.
3. IF an authenticated principal requests an operation not permitted for the principal's role, THEN THE API_Server SHALL respond with HTTP status 403.
4. WHERE an endpoint manages leases, THE Auth_Service SHALL permit access only to the SuperAdmin role.
5. WHEN a Company, Affiliate, or Merchant principal requests data, THE Tenant_Scope SHALL further restrict results to the records that principal owns within its Tenant, matching the current portal behavior.
6. THE API_Server SHALL require a valid Session_Token for every endpoint that reads or writes business data.

### Requirement 8: CRUD for all business entities

**User Story:** As an Admin, I want full create, read, update, and delete operations for every entity, so that I can manage my business through the API.

#### Acceptance Criteria

1. THE API_Server SHALL provide create, read, update, and delete operations for gateways, companies, affiliates, merchants, transactions, settlements, affiliate payments, and merchant payments.
2. WHEN a principal creates or updates a merchant, THE API_Server SHALL persist the merchant's nested banks, ATM cards, custom fields, and payment-gateway credentials in the same operation.
3. WHEN a principal deletes a merchant, THE API_Server SHALL delete the merchant's nested banks, ATM cards, and payment-gateway credentials.
4. IF a create or update request contains data that violates a required-field or referential constraint, THEN THE API_Server SHALL respond with HTTP status 400 and a descriptive error identifying the invalid field, and SHALL allow other concurrent requests to continue processing.
5. WHEN a principal reads a collection, THE API_Server SHALL return only records within the principal's Tenant and permitted scope.
6. IF a principal requests records outside its permitted scope within its Tenant, THEN THE API_Server SHALL return an empty result set and SHALL log the attempt.
7. THE API_Server SHALL store merchant bank and payment-gateway credential secrets in a manner that is not exposed to roles outside the owning Tenant.

### Requirement 9: Server-side commission calculation

**User Story:** As an Admin, I want commissions computed by the backend, so that the frontend consumes consistent, authoritative computed values.

#### Acceptance Criteria

1. WHEN a transaction is read with its computed breakdown, THE Commission_Engine SHALL compute Admin Commission as the gateway commission percentage applied to the transaction amount.
2. THE Commission_Engine SHALL compute Company Net Income as Settlement Amount minus Admin Commission, and SHALL further subtract transaction and other charges when the charge bearer is Company.
3. THE Commission_Engine SHALL compute Admin Net Income as Admin Commission minus the merchant or affiliate commission, and SHALL further subtract transaction and other charges when the charge bearer is Admin.
4. WHERE a merchant is assigned to an affiliate, THE Commission_Engine SHALL pay commission to the affiliate using the affiliate's commission percentage and base, and SHALL NOT report a separate merchant commission for that transaction.
5. WHERE a merchant is not assigned to an affiliate, THE Commission_Engine SHALL pay commission to the merchant using the merchant's commission percentage and base.
6. THE Commission_Engine SHALL select the commission base amount as the transaction amount when the configured base is "Transaction Amount" and as the settlement amount when the configured base is "Settlement Amount".
7. THE Commission_Engine SHALL produce results equivalent to the current frontend `calc.js` computation for the same inputs.

### Requirement 10: Ledger aggregation

**User Story:** As an Admin, Company, Affiliate, or Merchant, I want ledger balances computed by the backend, so that receivable, earned, and paid figures are accurate and consistent.

#### Acceptance Criteria

1. WHEN a company ledger is requested, THE Report_Service SHALL compute receivable as the sum of Company Net Income across the company's transactions, paid as the sum of settlement payments to the company, and balance as receivable minus paid.
2. WHEN an affiliate ledger is requested, THE Report_Service SHALL compute earned as the sum of commission across the affiliate's merchants' transactions, paid as the sum of affiliate payments, and balance as earned minus paid.
3. WHEN a merchant ledger is requested for a direct merchant, THE Report_Service SHALL compute earned as the sum of merchant commission across the merchant's transactions, paid as the sum of merchant payments, and balance as earned minus paid.
4. WHERE a merchant is assigned to an affiliate, THE Report_Service SHALL NOT compute merchant ledger values for that merchant.
5. THE Report_Service SHALL compute every ledger using only records within the requesting principal's Tenant and permitted scope.

### Requirement 11: Reports and exports

**User Story:** As an Admin, I want company, merchant, affiliate, gateway, settlement, and outstanding reports with date filtering and exports, so that I can analyze and share business data.

#### Acceptance Criteria

1. THE Report_Service SHALL produce company-wise, merchant-wise, affiliate-wise, gateway-wise, settlement-wise, and outstanding-wise reports.
2. WHEN a report request includes a start date and end date, THE Report_Service SHALL include only records whose date falls within the inclusive range.
3. WHEN a client requests a report export in CSV format, THE Report_Service SHALL return the report data as a CSV file.
4. WHEN a client requests a report export in PDF format, THE Report_Service SHALL return the report data as a PDF file.
5. WHEN a client requests exports in multiple formats and one format fails, THE Report_Service SHALL return each format that succeeds independently of the format that fails.
6. THE Report_Service SHALL compute every report using only records within the requesting principal's Tenant and permitted scope.

### Requirement 12: SuperAdmin role and own business

**User Story:** As a SuperAdmin, I want to run my own business with full Admin functionality, so that I operate as a tenant in addition to managing leases.

#### Acceptance Criteria

1. THE API_Server SHALL grant the SuperAdmin role all operations available to the Admin role within the SuperAdmin's own Tenant.
2. THE Migration_Service SHALL create a SuperAdmin demo account with documented login credentials.
3. WHEN the SuperAdmin manages its own companies, affiliates, merchants, gateways, transactions, settlements, ledgers, and reports, THE Tenant_Scope SHALL constrain those operations to the SuperAdmin's own Tenant.
4. THE API_Server SHALL map the existing demo Admin account to its own dedicated Tenant, separate from the SuperAdmin's Tenant.

### Requirement 13: Lease creation and tenure

**User Story:** As a SuperAdmin, I want to create an Admin account with a leased platform instance and a defined tenure, so that the new Admin can operate an isolated business for a set period.

#### Acceptance Criteria

1. WHEN a SuperAdmin creates a Lease with a new Admin user id, password, start date, and expiry date, THE Lease_Manager SHALL create a new Admin account, a new dedicated Tenant, and a Lease record linking them.
2. WHEN the Lease_Manager creates a leased Admin account, THE Auth_Service SHALL store the account password as a one-way hash.
3. WHEN a Lease is created, THE Lease_Manager SHALL set the initial Lease_Status to Active.
4. IF a Lease creation request fails any field validation, including an expiry date that is not after the start date, THEN THE Lease_Manager SHALL respond with HTTP status 400 and a descriptive error.
5. IF a Lease creation request specifies a user id that already exists, THEN THE Lease_Manager SHALL respond with HTTP status 409 and a descriptive error.
6. IF a Lease creation request contains both a field validation failure and a duplicate user id, THEN THE Lease_Manager SHALL respond with HTTP status 400.
7. WHEN a Lease is created, THE Tenant associated with the new leased Admin SHALL contain no business records from any other Tenant.

### Requirement 14: Lease expiry enforcement

**User Story:** As a SuperAdmin, I want a leased Admin's access disabled when the tenure ends, so that access aligns with the agreed lease period.

#### Acceptance Criteria

1. WHILE the current date is after a Lease's expiry date, THE Lease_Manager SHALL treat the Lease_Status as Expired.
2. IF a leased Admin attempts to authenticate while the Lease_Status is Expired, THEN THE Auth_Service SHALL deny the login with HTTP status 403 and a message indicating the lease has expired.
3. IF a leased Admin presents a Session_Token while the Lease_Status is Expired, THEN THE API_Server SHALL respond with HTTP status 403 and SHALL NOT perform the requested operation.
4. WHEN a Lease becomes Expired, THE Lease_Manager SHALL retain the Tenant's data without deletion.
5. WHEN a Lease_Status changes to Expired, THE API_Server SHALL continue to allow the SuperAdmin to view and manage that Lease.

### Requirement 15: Lease management operations

**User Story:** As a SuperAdmin, I want to list, extend, suspend, and revoke leases, so that I can administer tenant access over time.

#### Acceptance Criteria

1. WHEN a SuperAdmin requests the lease list, THE Lease_Manager SHALL return all leases with their Admin user id, Tenant, start date, expiry date, and Lease_Status.
2. WHEN a SuperAdmin extends a Lease with a new expiry date later than the current expiry date, THE Lease_Manager SHALL update the expiry date and set the Lease_Status to Active.
3. WHEN a SuperAdmin suspends an Active Lease, THE Lease_Manager SHALL set the Lease_Status to Suspended.
4. IF a leased Admin attempts to authenticate or presents a Session_Token while the Lease_Status is Suspended, THEN THE Auth_Service SHALL deny access with HTTP status 403.
5. WHEN a SuperAdmin reactivates a Suspended Lease whose expiry date is in the future, THE Lease_Manager SHALL set the Lease_Status to Active.
6. WHEN a SuperAdmin revokes a Lease, THE Lease_Manager SHALL set the Lease_Status to Revoked and SHALL deny all access for the associated leased Admin.
7. IF a non-SuperAdmin principal requests any lease management operation, THEN THE API_Server SHALL respond with HTTP status 403.

### Requirement 16: SuperAdmin-only lease navigation

**User Story:** As a SuperAdmin, I want a dedicated navigation entry to manage leases, so that I can reach lease administration without exposing it to normal Admins.

#### Acceptance Criteria

1. WHERE the authenticated role is SuperAdmin, THE Web_Client SHALL display a lease management navigation entry positioned below the existing "Reports" navigation item.
2. WHERE the authenticated role is Admin, Company, Affiliate, or Merchant, THE Web_Client SHALL NOT display the lease management navigation entry.
3. WHEN a SuperAdmin selects the lease management navigation entry, THE Web_Client SHALL display the lease management screen for listing and managing leases.
4. THE Web_Client SHALL display all existing Admin, Company, Affiliate, and Merchant portal navigation entries unchanged for their respective roles.

### Requirement 17: Frontend data-layer migration

**User Story:** As a user, I want the existing screens to work against the backend, so that all current behavior is preserved without localStorage.

#### Acceptance Criteria

1. THE Web_Client SHALL retrieve and persist all business data through the API_Server instead of browser localStorage.
2. THE Web_Client SHALL store the Session_Token on login and include it on every authenticated API request.
3. WHEN the Web_Client displays commission, ledger, or report values, THE Web_Client SHALL use the values computed by the API_Server.
4. THE Web_Client SHALL preserve all existing Admin, Company, Affiliate, and Merchant screens and their current behavior.
5. WHEN a user logs out, THE Web_Client SHALL discard the stored Session_Token.
6. IF an API request returns HTTP status 401, THEN THE Web_Client SHALL return the user to the login screen.

### Requirement 18: API error handling

**User Story:** As a frontend developer, I want consistent, descriptive API errors, so that the client can present meaningful feedback.

#### Acceptance Criteria

1. WHEN the API_Server rejects a request, THE API_Server SHALL respond with a structured error body containing a machine-readable code and a human-readable message.
2. IF a request references an entity that does not exist within the principal's Tenant, THEN THE API_Server SHALL respond with HTTP status 404.
3. IF request validation fails, THEN THE API_Server SHALL respond with HTTP status 400 and identify the invalid fields.
4. IF an unexpected server error occurs, THEN THE API_Server SHALL respond with HTTP status 500 and SHALL NOT expose internal stack traces or secret values in the response body.
5. WHEN the API_Server returns an error, THE API_Server SHALL use HTTP status codes consistent with the error condition.
