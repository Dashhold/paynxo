package crud

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/repo"
	"pgcs/backend/internal/service"
)

// CompanyService is the tenant-scoped CRUD service for companies. A company
// owns a set of CompanyGateway assignments (gateway + commission % + charge
// bearer) that must be persisted together with the company (Req 8.2). Reads
// preload the assignments; creates insert them in one operation; updates fully
// replace them so removing an assignment from the request removes it from the
// database.
//
// Tenant isolation is enforced through repo.ScopeTenant on every read, update,
// and delete, exactly as the generic Service does, while the nested-association
// handling that the generic Service cannot express lives here.
type CompanyService struct {
	db *gorm.DB
}

// NewCompanyService constructs the company CRUD service bound to a database
// handle.
func NewCompanyService(db *gorm.DB) *CompanyService {
	return &CompanyService{db: db}
}

// List returns the companies visible to the principal within its tenant scope,
// each with its gateway assignments preloaded (Req 8.5).
func (s *CompanyService) List(p service.Principal) ([]model.Company, error) {
	var out []model.Company
	if err := s.db.Scopes(repo.ScopeTenant(p)).Preload("Gateways").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns the company with the given id within the principal's scope, with
// its gateway assignments preloaded. A company that does not exist within the
// scope is reported as apierr.ErrNotFound so cross-tenant reads are
// indistinguishable from "does not exist" (Req 4.3, 18.2).
func (s *CompanyService) Get(p service.Principal, id string) (*model.Company, error) {
	var c model.Company
	err := s.db.Scopes(repo.ScopeTenant(p)).Preload("Gateways").Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Create validates the company, stamps the tenant on the company and each of
// its gateway assignments, inserts them together, and creates the company's
// portal login account — all in one transaction (Req 4.2, 8.2). Missing ids
// are generated.
func (s *CompanyService) Create(p service.Principal, c *model.Company) error {
	if err := validateCompany(c); err != nil {
		return err
	}
	if c.ID == "" {
		c.ID = GenID()
	}
	c.TenantID = p.TenantID
	stampAssignments(c, p.TenantID)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(c).Error; err != nil {
			return err
		}
		return upsertPortalAccount(tx, p.TenantID, service.RoleCompany, "Company", c.ID, c.UserID, c.Password, c.Name)
	})
	c.Password = "" // never echo the submitted password back
	return conflictIfUnique(err)
}

// Update validates the company, confirms it exists within the principal's
// scope, then persists the company and fully replaces its gateway assignments
// in a single transaction. Replacing (delete existing, then full-session save
// of the supplied set) means an assignment dropped from the request is removed
// rather than orphaned (Req 8.2). Updating a company outside the principal's
// scope returns apierr.ErrNotFound and changes nothing (Req 4.3).
func (s *CompanyService) Update(p service.Principal, c *model.Company) error {
	if err := validateCompany(c); err != nil {
		return err
	}
	if _, err := s.Get(p, c.ID); err != nil {
		return err
	}
	c.TenantID = p.TenantID
	stampAssignments(c, p.TenantID)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Remove the company's existing assignments within its tenant, then
		// recreate the supplied set via a full-session save so updates replace
		// assignments instead of merging them.
		if err := tx.Where("company_id = ? AND tenant_id = ?", c.ID, p.TenantID).
			Delete(&model.CompanyGateway{}).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(c).Error; err != nil {
			return err
		}
		// Keep the portal login in sync: update user id/name always, password
		// only when a new one was supplied.
		return upsertPortalAccount(tx, p.TenantID, service.RoleCompany, "Company", c.ID, c.UserID, c.Password, c.Name)
	})
	c.Password = ""
	return conflictIfUnique(err)
}

// Delete removes the company within the principal's scope together with its
// gateway assignments, so no orphaned assignments remain. A company outside the
// scope yields apierr.ErrNotFound (Req 4.3).
func (s *CompanyService) Delete(p service.Principal, id string) error {
	if _, err := s.Get(p, id); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("company_id = ? AND tenant_id = ?", id, p.TenantID).
			Delete(&model.CompanyGateway{}).Error; err != nil {
			return err
		}
		if err := deletePortalAccount(tx, p.TenantID, "Company", id); err != nil {
			return err
		}
		return tx.Scopes(repo.ScopeTenant(p)).Where("id = ?", id).
			Delete(&model.Company{}).Error
	})
}

// stampAssignments fills the tenant id, company id, and any missing assignment
// id on each of the company's gateway assignments so they persist consistently
// with their parent.
func stampAssignments(c *model.Company, tenantID string) {
	for i := range c.Gateways {
		if c.Gateways[i].ID == "" {
			c.Gateways[i].ID = GenID()
		}
		c.Gateways[i].TenantID = tenantID
		c.Gateways[i].CompanyID = c.ID
	}
}

// validateCompany enforces the company's required fields, returning a 400
// validation error that identifies the offending field (Req 8.4).
func validateCompany(c *model.Company) error {
	if strings.TrimSpace(c.Name) == "" {
		return apierr.ValidationField("name", "company name is required")
	}
	return nil
}
