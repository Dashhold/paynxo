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

// MerchantService is the tenant-scoped CRUD service for merchants. A merchant
// owns deeply nested data: banks (each owning ATM cards and polymorphic custom
// fields) and payment-gateway credentials (each owning polymorphic custom
// fields). All of it must be persisted together with the merchant (Req 8.2)
// and deleted together with it (Req 8.3).
//
// Like CompanyService, tenant isolation is enforced through repo.ScopeTenant on
// every read, update, and delete, while the multi-level association handling
// that the generic Service cannot express lives here. Creates and updates run
// in a single transaction so the merchant and all its nested rows commit or
// roll back together; deletes cascade the nested rows explicitly (rather than
// relying on database foreign-key cascade, which AutoMigrate does not
// guarantee) inside a transaction.
type MerchantService struct {
	db *gorm.DB
}

// NewMerchantService constructs the merchant CRUD service bound to a database
// handle.
func NewMerchantService(db *gorm.DB) *MerchantService {
	return &MerchantService{db: db}
}

// preloadNested chains the preloads that load a merchant's full nested tree:
// banks with their ATM cards and custom fields, and payment-gateway credentials
// with their custom fields.
func preloadNested(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Banks.AtmCards").
		Preload("Banks.Custom").
		Preload("PaymentGateways.Custom")
}

// List returns the merchants visible to the principal within its tenant scope,
// each with its full nested tree preloaded (Req 8.5).
func (s *MerchantService) List(p service.Principal) ([]model.Merchant, error) {
	var out []model.Merchant
	q := preloadNested(s.db.Scopes(repo.ScopeTenant(p)))
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns the merchant with the given id within the principal's scope, with
// its full nested tree preloaded. A merchant that does not exist within the
// scope is reported as apierr.ErrNotFound so cross-tenant reads are
// indistinguishable from "does not exist" (Req 4.3, 18.2).
func (s *MerchantService) Get(p service.Principal, id string) (*model.Merchant, error) {
	var m model.Merchant
	q := preloadNested(s.db.Scopes(repo.ScopeTenant(p)))
	err := q.Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apierr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create validates the merchant, stamps the tenant on the merchant and every
// nested record (banks, ATM cards, gateway credentials, custom fields),
// generates ids where missing, and inserts the whole tree in one transaction
// (Req 4.2, 8.2). GORM persists the nested associations recursively as part of
// the create.
func (s *MerchantService) Create(p service.Principal, m *model.Merchant) error {
	if err := validateMerchant(m); err != nil {
		return err
	}
	if m.ID == "" {
		m.ID = GenID()
	}
	m.TenantID = p.TenantID
	stampMerchantNested(m, p.TenantID)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		return upsertPortalAccount(tx, p.TenantID, service.RoleMerchant, "Merchant", m.ID, m.UserID, m.Password, m.Name)
	})
	m.Password = "" // never echo the submitted password back
	return conflictIfUnique(err)
}

// Update validates the merchant, confirms it exists within the principal's
// scope, then persists the merchant and fully replaces its nested tree in a
// single transaction. Replacing (delete the existing nested rows for the
// merchant, then full-session save of the supplied tree) means a bank, card,
// credential, or custom field dropped from the request is removed rather than
// orphaned (Req 8.2). Updating a merchant outside the principal's scope returns
// apierr.ErrNotFound and changes nothing (Req 4.3).
func (s *MerchantService) Update(p service.Principal, m *model.Merchant) error {
	if err := validateMerchant(m); err != nil {
		return err
	}
	if _, err := s.Get(p, m.ID); err != nil {
		return err
	}
	m.TenantID = p.TenantID
	stampMerchantNested(m, p.TenantID)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := cascadeDeleteMerchantNested(tx, p.TenantID, m.ID); err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(m).Error; err != nil {
			return err
		}
		return upsertPortalAccount(tx, p.TenantID, service.RoleMerchant, "Merchant", m.ID, m.UserID, m.Password, m.Name)
	})
	m.Password = ""
	return conflictIfUnique(err)
}

// Delete removes the merchant within the principal's scope together with its
// entire nested tree, so no orphaned banks, ATM cards, credentials, or custom
// fields remain (Req 8.3). A merchant outside the scope yields
// apierr.ErrNotFound (Req 4.3). The cascade is performed explicitly within a
// transaction rather than relying on database FK cascade.
func (s *MerchantService) Delete(p service.Principal, id string) error {
	if _, err := s.Get(p, id); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := cascadeDeleteMerchantNested(tx, p.TenantID, id); err != nil {
			return err
		}
		if err := deletePortalAccount(tx, p.TenantID, "Merchant", id); err != nil {
			return err
		}
		return tx.Scopes(repo.ScopeTenant(p)).Where("id = ?", id).
			Delete(&model.Merchant{}).Error
	})
}

// cascadeDeleteMerchantNested removes every nested record owned by the merchant
// within the given tenant: custom fields owned by the merchant's banks and
// credentials, the ATM cards owned by its banks, then the banks and credentials
// themselves. Children are deleted before parents so no row is orphaned, and
// every delete is constrained to the tenant so the cascade can never reach
// another tenant's data (Req 8.3).
func cascadeDeleteMerchantNested(tx *gorm.DB, tenantID, merchantID string) error {
	var bankIDs []string
	if err := tx.Model(&model.MerchantBank{}).
		Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID).
		Pluck("id", &bankIDs).Error; err != nil {
		return err
	}
	var credIDs []string
	if err := tx.Model(&model.MerchantGatewayCredential{}).
		Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID).
		Pluck("id", &credIDs).Error; err != nil {
		return err
	}

	// Custom fields are polymorphic: they are owned by either a bank or a
	// credential. Delete those whose owner is one of this merchant's banks or
	// credentials, scoped to the tenant.
	ownerIDs := make([]string, 0, len(bankIDs)+len(credIDs))
	ownerIDs = append(ownerIDs, bankIDs...)
	ownerIDs = append(ownerIDs, credIDs...)
	if len(ownerIDs) > 0 {
		if err := tx.Where("tenant_id = ? AND owner_id IN ?", tenantID, ownerIDs).
			Delete(&model.CustomField{}).Error; err != nil {
			return err
		}
	}

	// ATM cards belong to the merchant's banks.
	if len(bankIDs) > 0 {
		if err := tx.Where("tenant_id = ? AND merchant_bank_id IN ?", tenantID, bankIDs).
			Delete(&model.AtmCard{}).Error; err != nil {
			return err
		}
	}

	if err := tx.Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID).
		Delete(&model.MerchantBank{}).Error; err != nil {
		return err
	}
	if err := tx.Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID).
		Delete(&model.MerchantGatewayCredential{}).Error; err != nil {
		return err
	}
	return nil
}

// stampMerchantNested fills the tenant id, owning foreign keys, and any missing
// ids on every nested record so the tree persists consistently with its parent
// merchant (Req 8.2). The polymorphic OwnerID/OwnerType on custom fields is set
// by GORM during the association save; here only the tenant id and id are
// stamped.
func stampMerchantNested(m *model.Merchant, tenantID string) {
	for i := range m.Banks {
		b := &m.Banks[i]
		if b.ID == "" {
			b.ID = GenID()
		}
		b.TenantID = tenantID
		b.MerchantID = m.ID
		for j := range b.AtmCards {
			c := &b.AtmCards[j]
			if c.ID == "" {
				c.ID = GenID()
			}
			c.TenantID = tenantID
			c.MerchantBankID = b.ID
		}
		stampCustomFields(b.Custom, tenantID)
	}
	for i := range m.PaymentGateways {
		g := &m.PaymentGateways[i]
		if g.ID == "" {
			g.ID = GenID()
		}
		g.TenantID = tenantID
		g.MerchantID = m.ID
		stampCustomFields(g.Custom, tenantID)
	}
}

// stampCustomFields stamps the tenant id and a generated id (when missing) on
// each polymorphic custom field.
func stampCustomFields(fields []model.CustomField, tenantID string) {
	for i := range fields {
		if fields[i].ID == "" {
			fields[i].ID = GenID()
		}
		fields[i].TenantID = tenantID
	}
}

// validateMerchant enforces the merchant's required fields, returning a 400
// validation error that identifies the offending field (Req 8.4).
func validateMerchant(m *model.Merchant) error {
	if strings.TrimSpace(m.Name) == "" {
		return apierr.ValidationField("name", "merchant name is required")
	}
	return nil
}
