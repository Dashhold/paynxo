package model

import "time"

// Tenant is the isolation boundary that owns a set of business entities. Every
// Admin (including the SuperAdmin's own business) maps to exactly one Tenant
// (Req 3.6, 4.4).
type Tenant struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"` // "superadmin-own" | "admin-own" | "leased"
	CreatedAt time.Time `json:"createdAt"`
}

// Account represents a login principal. SuperAdmin/Admin accounts map to one
// Tenant; Company/Affiliate/Merchant portal logins also reference their owning
// business entity via OwnerType/OwnerID, replacing the plaintext
// userId/password fields stored on those entities in seed.js. Passwords are
// stored only as one-way bcrypt hashes, never plaintext (Req 5.5, 6.4).
// For SuperAdmin convenience, the plaintext password is also stored encrypted
// in PasswordEncrypted (reversible encryption) so it can be viewed later.
type Account struct {
	ID                string `gorm:"primaryKey" json:"id"`
	UserID            string `gorm:"uniqueIndex;not null" json:"userId"`
	PasswordHash      string `gorm:"not null" json:"-"` // bcrypt; never plaintext
	PasswordEncrypted string `json:"-"` // AES encrypted plaintext for SuperAdmin viewing
	Role              string `gorm:"not null" json:"role"` // SuperAdmin|Admin|Company|Affiliate|Merchant
	TenantID          string `gorm:"index;not null" json:"tenantId"`
	OwnerType         string `json:"ownerType"` // "", Company|Affiliate|Merchant
	OwnerID           string `json:"ownerId"`   // entity id for portal accounts
	Name              string `json:"name"`
}
