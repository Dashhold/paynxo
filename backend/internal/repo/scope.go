package repo

import (
	"gorm.io/gorm"

	"pgcs/backend/internal/service"
)

// Entity is the constraint satisfied by every tenant-scoped business entity.
// It is implemented by *T for any model struct that embeds model.TenantBase
// (Gateway, Company, Affiliate, Merchant, Transaction, Settlement,
// AffiliatePayment, MerchantPayment, ...). The generic Repository uses it to
// read a record's id and to stamp its owning tenant on writes.
type Entity interface {
	GetID() string
	GetTenantID() string
	SetTenantID(string)
}

// ScopeTenant returns a GORM scope that constrains a query to the records the
// principal is allowed to see and modify. It is applied to every
// business-entity read, update, and delete so isolation cannot be bypassed by
// forgetting a WHERE clause (Req 4.1, 4.5, 7.5).
//
// All principals are restricted to their own tenant. Portal roles
// (Company/Affiliate/Merchant) are further restricted to the records they own
// within that tenant, matching the existing portal behavior (Req 7.5):
//
//   - Company   -> company_id = OwnerID
//   - Affiliate -> affiliate_id = OwnerID
//   - Merchant  -> merchant_id = OwnerID OR id = OwnerID
//     (covers both records that reference the merchant and the merchant row
//     itself)
//
// SuperAdmin and Admin apply only the tenant filter. The owner predicates are
// combined with explicit parentheses so the OR in the Merchant case cannot
// widen the result beyond the tenant.
func ScopeTenant(p service.Principal) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch p.Role {
		case service.RoleCompany:
			return db.Where("tenant_id = ? AND company_id = ?", p.TenantID, p.OwnerID)
		case service.RoleAffiliate:
			return db.Where("tenant_id = ? AND affiliate_id = ?", p.TenantID, p.OwnerID)
		case service.RoleMerchant:
			return db.Where("tenant_id = ? AND (merchant_id = ? OR id = ?)", p.TenantID, p.OwnerID, p.OwnerID)
		default: // SuperAdmin, Admin: tenant filter only.
			return db.Where("tenant_id = ?", p.TenantID)
		}
	}
}
