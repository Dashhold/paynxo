// This file implements the Lease_Manager operations that build on the status
// model defined in lease.go: creating a lease (with its dedicated Tenant and
// leased Admin Account), listing leases with their *effective* status, and the
// extend / suspend / reactivate / revoke administrative transitions
// (Req 13, 14.4, 15.1, 15.2, 15.3, 15.5, 15.6).
//
// Stored vs. derived status: every transition writes only administrative
// intent (Active | Suspended | Revoked) to model.Lease.Status. The Expired
// status is never persisted; List derives it through EffectiveStatus using the
// manager's now func so tenure enforcement needs no scheduled job (Req 14.1).
package lease

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/crypto"
)

// CreateLeaseInput is the request to provision a new leased Admin: the Admin's
// login user id and password plus the tenure bounds (Req 13.1).
type CreateLeaseInput struct {
	AdminUserID string
	AdminName   string
	Password    string
	StartDate   time.Time
	ExpiryDate  time.Time
}

// LeaseView is the read projection returned by List: the lease's identity,
// tenure, and *effective* Lease_Status resolved at read time (Req 15.1).
type LeaseView struct {
	ID          string      `json:"id"`
	AdminUserID string      `json:"adminUserId"`
	AdminName   string      `json:"adminName"`
	TenantID    string      `json:"tenantId"`
	AccountID   string      `json:"accountId"`
	StartDate   time.Time   `json:"startDate"`
	ExpiryDate  time.Time   `json:"expiryDate"`
	Status      LeaseStatus `json:"status"`
}

// LeaseManager is the behavior the lease HTTP handlers depend on. *Manager is
// the concrete implementation. All operations are SuperAdmin-only; that
// authorization is enforced by the handlers (Req 7.4, 15.7), not here.
type LeaseManager interface {
	// Create provisions a new Tenant, leased Admin Account, and Lease (Req 13).
	Create(in CreateLeaseInput) (model.Lease, error)
	// List returns every lease with its effective status (Req 15.1).
	List() ([]LeaseView, error)
	// Extend moves the expiry later and re-activates the lease (Req 15.2).
	Extend(id string, newExpiry time.Time) (model.Lease, error)
	// Suspend pauses an active lease (Req 15.3).
	Suspend(id string) (model.Lease, error)
	// Reactivate clears a suspension (Req 15.5).
	Reactivate(id string) (model.Lease, error)
	// Revoke permanently ends a lease (Req 15.6).
	Revoke(id string) (model.Lease, error)
}

// Manager implements LeaseManager over a GORM database.
type Manager struct {
	db *gorm.DB
	// now returns the current time; overridable in tests for deterministic
	// effective-status derivation, consistent with the Auth_Service pattern.
	now func() time.Time
}

// Ensure *Manager satisfies the interface.
var _ LeaseManager = (*Manager)(nil)

// NewManager constructs a Lease_Manager bound to a database handle.
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db, now: time.Now}
}

// Create validates the input, rejects a duplicate user id, then provisions —
// in a single transaction — a new leased Tenant, a bcrypt-hashed Admin
// Account, and an Active Lease linking them (Req 13.1, 13.2, 13.3).
//
// Error precedence is deliberate: field validation runs first, so a request
// that both fails validation AND names an existing user id returns a 400
// validation error rather than a 409 conflict (Req 13.4, 13.6). Only once the
// input is well-formed is the duplicate-user-id check performed, which returns
// 409 (Req 13.5).
//
// The transaction guarantees the new tenant is created fresh and in isolation;
// it is populated with no records from any other tenant (Req 13.7, 14.4).
func (m *Manager) Create(in CreateLeaseInput) (model.Lease, error) {
	// Validation precedence: a malformed request is a 400 regardless of any
	// duplicate user id (Req 13.6).
	if err := validateCreateInput(in); err != nil {
		return model.Lease{}, err
	}

	// Duplicate user id -> 409, but only for an otherwise valid request
	// (Req 13.5). Case-insensitive so ids cannot collide by case.
	var count int64
	if err := m.db.Model(&model.Account{}).
		Where("LOWER(user_id) = LOWER(?)", in.AdminUserID).
		Count(&count).Error; err != nil {
		return model.Lease{}, err
	}
	if count > 0 {
		return model.Lease{}, apierr.Conflict("a user with this id already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.Lease{}, err
	}

	// Also encrypt the plaintext password for SuperAdmin viewing
	encrypted, err := crypto.EncryptPassword(in.Password)
	if err != nil {
		return model.Lease{}, err
	}

	now := m.now()
	adminName := strings.TrimSpace(in.AdminName)
	if adminName == "" {
		adminName = in.AdminUserID
	}
	tenant := model.Tenant{
		ID:        genID(),
		Name:      adminName,
		Kind:      "leased",
		CreatedAt: now,
	}
	account := model.Account{
		ID:                genID(),
		UserID:            in.AdminUserID,
		PasswordHash:      string(hash),
		PasswordEncrypted: encrypted,
		Role:              service.RoleAdmin,
		TenantID:          tenant.ID,
		Name:              adminName,
	}
	lease := model.Lease{
		ID:          genID(),
		TenantID:    tenant.ID,
		AccountID:   account.ID,
		AdminUserID: in.AdminUserID,
		AdminName:   adminName,
		StartDate:   in.StartDate,
		ExpiryDate:  in.ExpiryDate,
		Status:      string(Active),
		CreatedAt:   now,
	}

	// All three rows are created atomically: a failure on any of them rolls
	// back the others so a half-provisioned lease can never exist.
	if err := m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}
		return tx.Create(&lease).Error
	}); err != nil {
		if apierr.IsUniqueViolation(err) {
			return model.Lease{}, apierr.Conflict("a user with this id already exists")
		}
		return model.Lease{}, err
	}

	return lease, nil
}

// List returns every lease with its identity, tenure, and effective
// Lease_Status. The stored intent is resolved through EffectiveStatus against
// the current time so an out-of-tenure Active lease is reported as Expired
// (Req 15.1, 14.1).
func (m *Manager) List() ([]LeaseView, error) {
	var leases []model.Lease
	if err := m.db.Order("created_at asc").Find(&leases).Error; err != nil {
		return nil, err
	}

	now := m.now()
	views := make([]LeaseView, 0, len(leases))
	for _, l := range leases {
		views = append(views, LeaseView{
			ID:          l.ID,
			AdminUserID: l.AdminUserID,
			AdminName:   l.AdminName,
			TenantID:    l.TenantID,
			AccountID:   l.AccountID,
			StartDate:   l.StartDate,
			ExpiryDate:  l.ExpiryDate,
			Status:      EffectiveStatus(l, now),
		})
	}
	return views, nil
}

// Extend sets a later expiry date and resets the stored intent to Active,
// bringing an expired or active lease (back) into tenure (Req 15.2). The new
// expiry must be strictly later than the current expiry, else 400 (the
// operation must actually extend the tenure).
func (m *Manager) Extend(id string, newExpiry time.Time) (model.Lease, error) {
	l, err := m.get(id)
	if err != nil {
		return model.Lease{}, err
	}
	if err := validateExtend(l.ExpiryDate, newExpiry); err != nil {
		return model.Lease{}, err
	}

	l.ExpiryDate = newExpiry
	l.Status = string(Active)
	if err := m.db.Save(&l).Error; err != nil {
		return model.Lease{}, err
	}
	return l, nil
}

// Suspend records the Suspended intent, pausing access while preserving the
// tenure (Req 15.3). Suspension overrides expiry for access purposes (Req 15.4).
func (m *Manager) Suspend(id string) (model.Lease, error) {
	return m.setStatus(id, Suspended)
}

// Reactivate clears a suspension by restoring the Active intent. It is only
// meaningful when the expiry is still in the future; once Active is stored, a
// past expiry resolves back to Expired through EffectiveStatus (Req 15.5).
func (m *Manager) Reactivate(id string) (model.Lease, error) {
	return m.setStatus(id, Active)
}

// Revoke records the terminal Revoked intent, permanently denying access for
// the associated leased Admin (Req 15.6).
func (m *Manager) Revoke(id string) (model.Lease, error) {
	return m.setStatus(id, Revoked)
}

// setStatus loads the lease, writes the given stored intent, and persists it.
// A missing lease id resolves to a 404 (Req 18.2).
func (m *Manager) setStatus(id string, status LeaseStatus) (model.Lease, error) {
	l, err := m.get(id)
	if err != nil {
		return model.Lease{}, err
	}
	l.Status = string(status)
	if err := m.db.Save(&l).Error; err != nil {
		return model.Lease{}, err
	}
	return l, nil
}

// get loads a lease by id, translating a missing row into a typed 404 so every
// operation on a non-existent lease returns not-found (Req 18.2).
func (m *Manager) get(id string) (model.Lease, error) {
	var l model.Lease
	err := m.db.Where("id = ?", id).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Lease{}, apierr.NotFound("lease not found")
	}
	if err != nil {
		return model.Lease{}, err
	}
	return l, nil
}

// validateCreateInput checks every required field on a create request and the
// tenure ordering, accumulating all offending fields into one 400 validation
// error (Req 13.4). It performs no I/O so it can run — and be tested — without
// a database, and it always runs before the duplicate-user-id check so a
// validation failure takes precedence over a conflict (Req 13.6).
func validateCreateInput(in CreateLeaseInput) error {
	fields := map[string]string{}
	if strings.TrimSpace(in.AdminUserID) == "" {
		fields["adminUserId"] = "admin user id is required"
	}
	if in.Password == "" {
		fields["password"] = "password is required"
	}
	if in.StartDate.IsZero() {
		fields["startDate"] = "start date is required"
	}
	switch {
	case in.ExpiryDate.IsZero():
		fields["expiryDate"] = "expiry date is required"
	case !in.ExpiryDate.After(in.StartDate):
		// Tenure must be a non-empty forward interval (Req 13.4).
		fields["expiryDate"] = "expiry date must be after the start date"
	}
	if len(fields) > 0 {
		return apierr.Validation("lease validation failed", fields)
	}
	return nil
}

// validateExtend enforces that an extension actually moves the expiry forward:
// the new expiry must be strictly later than the current one (Req 15.2).
func validateExtend(current, newExpiry time.Time) error {
	if !newExpiry.After(current) {
		return apierr.ValidationField("expiryDate", "new expiry date must be later than the current expiry date")
	}
	return nil
}

// genID returns a random, collision-resistant identifier for a new tenant,
// account, or lease row. It mirrors the id style used elsewhere in the backend
// without coupling this package to the CRUD layer.
func genID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		// rand.Read essentially never fails; avoid panicking in that case.
		return "id_" + hex.EncodeToString(b)
	}
	return "id_" + hex.EncodeToString(b)
}
