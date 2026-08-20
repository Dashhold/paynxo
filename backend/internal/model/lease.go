package model

import "time"

// Lease is a grant from a SuperAdmin to an Admin authorizing operation of a
// Tenant for a defined tenure. The stored Status holds the administrative
// intent (Active | Suspended | Revoked); the Expired status is derived at read
// time by comparing now against ExpiryDate.
type Lease struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"uniqueIndex" json:"tenantId"`
	AccountID   string    `gorm:"uniqueIndex" json:"accountId"`
	AdminUserID string    `json:"adminUserId"`
	AdminName   string    `json:"adminName"`
	StartDate   time.Time `json:"startDate"`
	ExpiryDate  time.Time `json:"expiryDate"`
	Status      string    `json:"status"` // Active | Suspended | Revoked (Expired is derived)
	CreatedAt   time.Time `json:"createdAt"`
}

// RevokedToken records a logged-out token's jti so the Auth_Service can reject
// it; ExpiresAt enables cleanup once the token would have expired anyway.
type RevokedToken struct {
	Jti       string    `gorm:"primaryKey" json:"jti"`
	ExpiresAt time.Time `json:"expiresAt"`
}
