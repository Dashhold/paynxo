package crud

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
)

// upsertPortalAccount creates or updates the login Account that backs a
// Company / Affiliate / Merchant portal entity, so the credentials captured on
// the entity form actually let that entity sign in.
//
// The account is linked to its business entity by (tenantID, ownerType,
// ownerID). On first create the user id and password are required; on update an
// empty password means "keep the current password" and an empty user id means
// "keep the current user id". The password is stored only as a bcrypt hash —
// plaintext is never persisted (Req 5.5, 6.4).
//
// User ids are globally unique (Account.UserID has a unique index), so a user
// id already taken by a different account yields a 409 conflict (Req 13.5
// semantics, applied to portal logins). All work runs on the supplied tx so it
// commits or rolls back together with the entity write.
func upsertPortalAccount(tx *gorm.DB, tenantID, role, ownerType, ownerID, userID, password, name string) error {
	userID = strings.TrimSpace(userID)

	var existing model.Account
	err := tx.Where("tenant_id = ? AND owner_type = ? AND owner_id = ?", tenantID, ownerType, ownerID).
		First(&existing).Error
	creating := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !creating {
		return err
	}

	// User-id uniqueness across all accounts (case-insensitive, excluding this
	// entity's own account when updating) so logins can never collide — not
	// even by case (e.g. "Admin" vs "admin").
	if userID != "" {
		q := tx.Model(&model.Account{}).Where("LOWER(user_id) = LOWER(?)", userID)
		if !creating {
			q = q.Where("id <> ?", existing.ID)
		}
		var count int64
		if err := q.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return apierr.Conflict("login user id is already taken")
		}
	}

	if creating {
		if userID == "" {
			return apierr.ValidationField("userId", "login user id is required")
		}
		if password == "" {
			return apierr.ValidationField("password", "login password is required")
		}
		hash, err := hashPassword(password)
		if err != nil {
			return err
		}
		return tx.Create(&model.Account{
			ID:           GenID(),
			UserID:       userID,
			PasswordHash: hash,
			Role:         role,
			TenantID:     tenantID,
			OwnerType:    ownerType,
			OwnerID:      ownerID,
			Name:         name,
		}).Error
	}

	// Update the existing portal account in place.
	if userID != "" {
		existing.UserID = userID
	}
	existing.Name = name
	if password != "" {
		hash, err := hashPassword(password)
		if err != nil {
			return err
		}
		existing.PasswordHash = hash
	}
	return tx.Save(&existing).Error
}

// deletePortalAccount removes the login Account backing a portal entity when
// that entity is deleted, so no orphaned login remains.
func deletePortalAccount(tx *gorm.DB, tenantID, ownerType, ownerID string) error {
	return tx.Where("tenant_id = ? AND owner_type = ? AND owner_id = ?", tenantID, ownerType, ownerID).
		Delete(&model.Account{}).Error
}

// conflictIfUnique converts a database unique-constraint violation (a duplicate
// login user id that raced past the application-level check) into a clean 409
// Conflict, leaving any other error unchanged.
func conflictIfUnique(err error) error {
	if apierr.IsUniqueViolation(err) {
		return apierr.Conflict("login user id is already taken")
	}
	return err
}

// hashPassword produces a bcrypt hash for a plaintext password so portal
// credentials are never stored in plaintext (Req 5.5, 6.4).
func hashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}
