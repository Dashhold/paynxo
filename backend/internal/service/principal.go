package service

// Principal identifies the authenticated caller for a request. It is produced
// by the Auth_Service from a validated Session_Token and consumed by the
// tenant-scope enforcement in the repository layer (see repo.ScopeTenant).
//
// It is defined in this leaf package (which imports no other internal package)
// rather than in a service subpackage so that both the repository layer and
// the auth/business subpackages can depend on it without creating an import
// cycle. The design refers to it as service.Principal, which this matches.
type Principal struct {
	// AccountID is the id of the Account this principal authenticated as.
	AccountID string
	// Role is one of the recognized roles (see the Role* constants).
	Role string
	// TenantID is the tenant the principal operates within. Every read and
	// write is constrained to this tenant (Req 4.1, 4.4, 4.5).
	TenantID string
	// OwnerType is "" for SuperAdmin/Admin, or Company|Affiliate|Merchant for
	// portal principals (Req 7.5).
	OwnerType string
	// OwnerID is the owning business-entity id for portal principals; empty for
	// SuperAdmin/Admin.
	OwnerID string
}

// Recognized roles (Req 7.1). These are the canonical string values stored on
// accounts and carried in the Session_Token.
const (
	RoleSuperAdmin = "SuperAdmin"
	RoleAdmin      = "Admin"
	RoleCompany    = "Company"
	RoleAffiliate  = "Affiliate"
	RoleMerchant   = "Merchant"
)
