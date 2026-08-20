// This file implements Migration_Service.SeedIfEmpty: idempotent bootstrap of
// the single root SuperAdmin account and its own tenant.
//
// The SuperAdmin is the only account created at startup. Every other account is
// created at runtime: Admins are provisioned by the SuperAdmin through leasing
// (each lease creates an Admin account + dedicated tenant), and the business
// entities an Admin manages live inside that Admin's tenant. No demo data and
// no demo logins are seeded.
//
// The SuperAdmin's credentials come from configuration (SUPERADMIN_USER_ID /
// SUPERADMIN_PASSWORD) and are stored only as a bcrypt hash; plaintext is never
// persisted (Req 5.5, 6.4).
//
// Idempotency (Req 5.4): the presence of the SuperAdmin tenant is the sentinel.
// If it already exists, SeedIfEmpty leaves every record unchanged and returns.
// The insert runs inside a single transaction so bootstrap is all-or-nothing.
package migration

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/crypto"
)

// timeNow is the time source for seeding; a package-level indirection keeps the
// CreatedAt stamps consistent across a single seed run and overridable in tests.
var timeNow = time.Now

// Fixed identifiers for the SuperAdmin's own tenant and account. The tenant id
// doubles as the idempotency sentinel: its presence means bootstrap has already
// run (Req 5.4).
const (
	superAdminTenantID  = "tenant-superadmin"
	superAdminAccountID = "acc-superadmin"
)

// SeedIfEmpty bootstraps the root SuperAdmin account and its tenant on first
// startup, or does nothing if bootstrap has already run.
//
// userID and password are the SuperAdmin's login credentials (from config). The
// password is hashed with bcrypt before storage. If the SuperAdmin tenant
// already exists, the function returns immediately, leaving all records
// unchanged (Req 5.4).
func SeedIfEmpty(db *gorm.DB, userID, password string) error {
	already, err := seedExists(db)
	if err != nil {
		return fmt.Errorf("check existing bootstrap: %w", err)
	}
	if already {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash superadmin password: %w", err)
	}

	// Also encrypt the plaintext password for SuperAdmin viewing
	encrypted, err := crypto.EncryptPassword(password)
	if err != nil {
		return fmt.Errorf("encrypt superadmin password: %w", err)
	}

	now := timeNow()
	tenant := model.Tenant{
		ID:        superAdminTenantID,
		Name:      "SuperAdmin Business",
		Kind:      "superadmin-own",
		CreatedAt: now,
	}
	account := model.Account{
		ID:                superAdminAccountID,
		UserID:            userID,
		PasswordHash:      string(hash),
		PasswordEncrypted: encrypted,
		Role:              service.RoleSuperAdmin,
		TenantID:          superAdminTenantID,
		Name:              "Super Administrator",
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}
		return tx.Create(&account).Error
	}); err != nil {
		return fmt.Errorf("bootstrap superadmin: %w", err)
	}
	return nil
}

// seedExists reports whether the SuperAdmin tenant (the bootstrap sentinel) is
// already present.
func seedExists(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Model(&model.Tenant{}).
		Where("id = ?", superAdminTenantID).
		Count(&count).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
