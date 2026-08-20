// Package crud implements the tenant-scoped create/read/update/delete services
// for the core business entities (gateways, companies, affiliates). Each
// service builds a generic, tenant-scoped repo.Repository from the request's
// authenticated principal so every read and write is constrained to the
// principal's tenant and owner scope (Req 4, 7.5, 8.5). Required-field
// validation runs before any write and surfaces a typed apierr.Validation so
// the HTTP layer renders a 400 identifying the offending field (Req 8.4).
//
// Most entities have no nested data and use the generic Service below.
// Companies carry CompanyGateway assignments and need association-aware
// persistence, so they have a dedicated CompanyService in this package
// (company.go).
package crud

import (
	"crypto/rand"
	"encoding/hex"

	"gorm.io/gorm"

	"pgcs/backend/internal/repo"
	"pgcs/backend/internal/service"
)

// Validator validates an entity before it is created or updated. It returns a
// typed apierr.Validation error (mapped to HTTP 400) when a required field or
// constraint is violated, and nil when the entity is acceptable.
type Validator[PT any] func(PT) error

// Service is a generic, tenant-scoped CRUD service over a single business
// entity type T (e.g. model.Gateway, model.Affiliate). It owns the database
// handle and a per-entity validator; every operation constructs a
// repo.Repository scoped to the caller's principal so tenant isolation is
// enforced by construction.
type Service[T any, PT interface {
	*T
	repo.Entity
}] struct {
	db       *gorm.DB
	validate Validator[PT]
}

// NewService constructs a generic CRUD service bound to a database handle and a
// validator. Pass a no-op validator (returning nil) for entities with no
// required fields.
func NewService[T any, PT interface {
	*T
	repo.Entity
}](db *gorm.DB, validate Validator[PT]) *Service[T, PT] {
	return &Service[T, PT]{db: db, validate: validate}
}

// repoFor builds a tenant-scoped repository for the given principal.
func (s *Service[T, PT]) repoFor(p service.Principal) *repo.Repository[T, PT] {
	return repo.New[T, PT](s.db, p)
}

// List returns every record visible to the principal within its tenant and
// owner scope (Req 8.5). Out-of-scope records are simply absent.
func (s *Service[T, PT]) List(p service.Principal) ([]T, error) {
	return s.repoFor(p).List()
}

// Get returns the record with the given id within the principal's scope, or
// apierr.ErrNotFound (404) when no in-scope record matches (Req 4.3, 18.2).
func (s *Service[T, PT]) Get(p service.Principal, id string) (PT, error) {
	return s.repoFor(p).Get(id)
}

// Create validates and inserts a new record, assigning its tenant from the
// principal (Req 4.2). A missing id is filled with a generated one so new
// records do not collide on an empty primary key.
func (s *Service[T, PT]) Create(p service.Principal, entity PT) error {
	if s.validate != nil {
		if err := s.validate(entity); err != nil {
			return err
		}
	}
	if entity.GetID() == "" {
		setID(entity, GenID())
	}
	return s.repoFor(p).Create(entity)
}

// Update validates and persists changes to an existing record. The record is
// resolved through the tenant/owner scope first, so updating a record outside
// the principal's scope returns apierr.ErrNotFound and changes nothing
// (Req 4.3).
func (s *Service[T, PT]) Update(p service.Principal, entity PT) error {
	if s.validate != nil {
		if err := s.validate(entity); err != nil {
			return err
		}
	}
	return s.repoFor(p).Update(entity)
}

// Delete removes the record with the given id within the principal's scope, or
// returns apierr.ErrNotFound when no in-scope record matches (Req 4.3).
func (s *Service[T, PT]) Delete(p service.Principal, id string) error {
	return s.repoFor(p).Delete(id)
}

// idSetter is implemented by entities that allow their primary key to be
// assigned. *T satisfies it via model.TenantBase's SetID below.
type idSetter interface {
	SetID(string)
}

// setID assigns a generated id to an entity that supports it. Entities that do
// not implement idSetter are left unchanged (their id, if required, must be
// supplied by the caller).
func setID(entity any, id string) {
	if s, ok := entity.(idSetter); ok {
		s.SetID(id)
	}
}

// GenID returns a random, collision-resistant identifier for a new record. New
// records use generated ids while seeded records keep their existing seed.js
// style ids.
func GenID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		// rand.Read essentially never fails; fall back to an empty-prefixed id
		// only to avoid panicking in the impossible case.
		return "id_" + hex.EncodeToString(b)
	}
	return "id_" + hex.EncodeToString(b)
}
