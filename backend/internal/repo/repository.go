package repo

import (
	"errors"

	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/service"
)

// Repository is a generic, tenant-scoped data-access helper for a single
// business-entity type T. Every read, update, and delete is run through
// ScopeTenant(principal) so a query can never cross the principal's tenant or
// owner scope. Create always stamps tenant_id from the principal.
//
// The two type parameters tie a value type T to its pointer type PT so the
// helpers can call the Entity methods (which have pointer receivers) while
// still scanning into addressable values of T.
type Repository[T any, PT interface {
	*T
	Entity
}] struct {
	db *gorm.DB
	p  service.Principal
}

// New constructs a Repository bound to a database handle and the authenticated
// principal. There is no unscoped constructor for business entities: a
// principal is always required, so isolation is enforced by construction.
func New[T any, PT interface {
	*T
	Entity
}](db *gorm.DB, p service.Principal) *Repository[T, PT] {
	return &Repository[T, PT]{db: db, p: p}
}

// List returns all records visible to the principal within its tenant and
// owner scope (Req 4.1, 7.5). Out-of-scope records are simply absent, so a
// portal role reading outside its ownership receives an empty slice.
func (r *Repository[T, PT]) List() ([]T, error) {
	var out []T
	if err := r.db.Scopes(ScopeTenant(r.p)).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns the record with the given id within the principal's scope. A row
// that does not exist, or that exists in another tenant or outside the
// principal's owner scope, is reported as apierr.ErrNotFound so cross-tenant
// reads are indistinguishable from "does not exist" and never disclose the
// record (Req 4.3, 18.2).
func (r *Repository[T, PT]) Get(id string) (PT, error) {
	var out T
	err := r.db.Scopes(ScopeTenant(r.p)).Where("id = ?", id).First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return PT(&out), nil
}

// Create inserts a new record, assigning its tenant from the principal and
// ignoring any client-supplied tenant id (Req 4.2). Owner fields (company_id,
// affiliate_id, merchant_id, ...) are the caller's responsibility.
func (r *Repository[T, PT]) Create(entity PT) error {
	entity.SetTenantID(r.p.TenantID)
	return r.db.Create(entity).Error
}

// Update persists changes to an existing record. The record is first resolved
// through the tenant/owner scope, so an attempt to update a record outside the
// principal's scope returns apierr.ErrNotFound and changes nothing (Req 4.3).
// The tenant id is re-stamped from the principal so it cannot be reassigned by
// a client.
func (r *Repository[T, PT]) Update(entity PT) error {
	if _, err := r.Get(entity.GetID()); err != nil {
		return err
	}
	entity.SetTenantID(r.p.TenantID)
	return r.db.Save(entity).Error
}

// Delete removes the record with the given id within the principal's scope. If
// no in-scope row matches, it returns apierr.ErrNotFound without disclosing
// whether the record exists in another tenant (Req 4.3).
func (r *Repository[T, PT]) Delete(id string) error {
	var zero T
	res := r.db.Scopes(ScopeTenant(r.p)).Where("id = ?", id).Delete(PT(&zero))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apierr.ErrNotFound
	}
	return nil
}
