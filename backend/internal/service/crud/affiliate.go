package crud

import (
	"errors"

	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/repo"
	"pgcs/backend/internal/service"
)

// AffiliateService is the tenant-scoped CRUD service for affiliates. Unlike the
// generic Service, it also provisions and maintains the affiliate's portal
// login account so a created affiliate can sign in (see accounts.go). Tenant
// isolation is enforced through repo.ScopeTenant on every read, update, and
// delete.
type AffiliateService struct {
	db *gorm.DB
}

// NewAffiliateService constructs the affiliate CRUD service bound to a database
// handle.
func NewAffiliateService(db *gorm.DB) *AffiliateService {
	return &AffiliateService{db: db}
}

// List returns the affiliates visible to the principal within its scope.
func (s *AffiliateService) List(p service.Principal) ([]model.Affiliate, error) {
	var out []model.Affiliate
	if err := s.db.Scopes(repo.ScopeTenant(p)).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns the affiliate with the given id within the principal's scope, or
// apierr.ErrNotFound when no in-scope record matches (Req 4.3, 18.2).
func (s *AffiliateService) Get(p service.Principal, id string) (*model.Affiliate, error) {
	var a model.Affiliate
	err := s.db.Scopes(repo.ScopeTenant(p)).Where("id = ?", id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Create validates the affiliate, inserts it, and provisions its portal login
// account, all in one transaction (Req 4.2, 8.4).
func (s *AffiliateService) Create(p service.Principal, a *model.Affiliate) error {
	if err := ValidateAffiliate(a); err != nil {
		return err
	}
	if a.ID == "" {
		a.ID = GenID()
	}
	a.TenantID = p.TenantID
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
		return upsertPortalAccount(tx, p.TenantID, service.RoleAffiliate, "Affiliate", a.ID, a.UserID, a.Password, a.Name)
	})
	a.Password = ""
	return conflictIfUnique(err)
}

// Update validates the affiliate, confirms it is in scope, persists the change,
// and keeps its portal login in sync (user id/name always, password only when a
// new one is supplied). An out-of-scope record yields apierr.ErrNotFound
// (Req 4.3).
func (s *AffiliateService) Update(p service.Principal, a *model.Affiliate) error {
	if err := ValidateAffiliate(a); err != nil {
		return err
	}
	if _, err := s.Get(p, a.ID); err != nil {
		return err
	}
	a.TenantID = p.TenantID
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(repo.ScopeTenant(p)).Save(a).Error; err != nil {
			return err
		}
		return upsertPortalAccount(tx, p.TenantID, service.RoleAffiliate, "Affiliate", a.ID, a.UserID, a.Password, a.Name)
	})
	a.Password = ""
	return conflictIfUnique(err)
}

// Delete removes the affiliate within the principal's scope together with its
// portal login account, or returns apierr.ErrNotFound (Req 4.3).
func (s *AffiliateService) Delete(p service.Principal, id string) error {
	if _, err := s.Get(p, id); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := deletePortalAccount(tx, p.TenantID, "Affiliate", id); err != nil {
			return err
		}
		return tx.Scopes(repo.ScopeTenant(p)).Where("id = ?", id).
			Delete(&model.Affiliate{}).Error
	})
}
