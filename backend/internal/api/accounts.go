package api

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/middleware"
	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
	"pgcs/backend/internal/service/crypto"
)

// accountHandlers serves the SuperAdmin-only account administration endpoints:
// GET /api/accounts/{id} to view an account's details, and PUT /api/accounts/{id}
// to edit the account's credentials (user id, password, name).
//
// Authorization to SuperAdmin is enforced by the route guard (Req 7.4), not here.
type accountHandlers struct {
	deps Deps
}

// newAccountHandlers builds the account handlers.
func newAccountHandlers(deps Deps) *accountHandlers {
	return &accountHandlers{deps: deps}
}

// get handles GET /api/accounts/{id}: returns an account's public details
// (userId, name, role, tenantId, ownerType/ownerId) and the decrypted password
// for SuperAdmin viewing. Password hashes are never disclosed (Req 5.5, 6.4).
// Returns 404 if the account does not exist.
func (h *accountHandlers) get(w http.ResponseWriter, r *http.Request) error {
	var acc model.Account
	err := h.deps.DB.Where("id = ?", r.PathValue("id")).First(&acc).Error
	if err != nil {
		return apierr.NotFound("account not found")
	}

	// Decrypt the password for SuperAdmin viewing
	password, err := crypto.DecryptPassword(acc.PasswordEncrypted)
	if err != nil {
		// If decryption fails, return empty password (backward compatibility)
		password = ""
	}

	middleware.WriteJSON(w, http.StatusOK, accountView{
		ID:        acc.ID,
		UserID:    acc.UserID,
		Name:      acc.Name,
		Role:      acc.Role,
		TenantID:  acc.TenantID,
		OwnerType: acc.OwnerType,
		OwnerID:   acc.OwnerID,
		Password:  password,
	})
	return nil
}

// update handles PUT /api/accounts/{id}: edits an account's user id, password,
// and name. All fields are optional; omitted fields are left unchanged. User id
// uniqueness is enforced (case-insensitive, returns 409 on conflict). Password
// is optional: if provided, it replaces the hash; if empty/omitted, the existing
// hash is kept. Returns 404 if the account does not exist.
func (h *accountHandlers) update(w http.ResponseWriter, r *http.Request) error {
	var req updateAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	// Load the existing account
	var acc model.Account
	err := h.deps.DB.Where("id = ?", r.PathValue("id")).First(&acc).Error
	if err != nil {
		return apierr.NotFound("account not found")
	}

	// Update user id if provided (enforce uniqueness, case-insensitive)
	newUserID := strings.TrimSpace(req.UserID)
	if newUserID != "" && strings.ToLower(newUserID) != strings.ToLower(acc.UserID) {
		var count int64
		if err := h.deps.DB.Model(&model.Account{}).
			Where("id != ? AND LOWER(user_id) = LOWER(?)", acc.ID, newUserID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return apierr.Conflict("a user with this id already exists")
		}
		acc.UserID = newUserID
	}

	// Update name if provided
	if strings.TrimSpace(req.Name) != "" {
		acc.Name = strings.TrimSpace(req.Name)
	}

	// Update password if provided (hash it and encrypt it)
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		acc.PasswordHash = string(hash)
		
		// Also encrypt the plaintext for SuperAdmin viewing
		encrypted, err := crypto.EncryptPassword(req.Password)
		if err != nil {
			return err
		}
		acc.PasswordEncrypted = encrypted
	}

	// Save the account
	if err := h.deps.DB.Save(&acc).Error; err != nil {
		// Check for db-level unique violation race
		if apierr.IsUniqueViolation(err) {
			return apierr.Conflict("a user with this id already exists")
		}
		return err
	}

	middleware.WriteJSON(w, http.StatusOK, accountView{
		ID:        acc.ID,
		UserID:    acc.UserID,
		Name:      acc.Name,
		Role:      acc.Role,
		TenantID:  acc.TenantID,
		OwnerType: acc.OwnerType,
		OwnerID:   acc.OwnerID,
		Password:  req.Password, // Return the new password if it was just set
	})
	return nil
}

// accountView is the read projection returned by GET /api/accounts/{id}.
// Password hashes are never included (Req 5.5, 6.4), but the decrypted plaintext
// password is returned for SuperAdmin viewing.
type accountView struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	TenantID  string `json:"tenantId"`
	OwnerType string `json:"ownerType"`
	OwnerID   string `json:"ownerId"`
	Password  string `json:"password"`
}

// updateAccountRequest is the PUT /api/accounts/{id} body. All fields are
// optional; omitted fields are left unchanged.
type updateAccountRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// registerAccountRoutes mounts the account administration endpoints (SuperAdmin-only).
func registerAccountRoutes(mux *http.ServeMux, d Deps) {
	ah := newAccountHandlers(d)
	superAdminOnly := RequireRoles(service.RoleSuperAdmin)

	mux.Handle("GET /api/accounts/{id}", protected(d.Auth, superAdminOnly(ah.get)))
	mux.Handle("PUT /api/accounts/{id}", protected(d.Auth, superAdminOnly(ah.update)))
}
