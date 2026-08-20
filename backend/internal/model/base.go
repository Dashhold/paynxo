// Package model defines the GORM data model structs and their relationships for
// all tenant-scoped business entities, mirroring the existing frontend data
// shapes from seed.js / calc.js. Implemented in task 2.1.
package model

import "time"

// TenantBase is embedded by every business entity. It carries the string
// primary key (preserving the existing seed.js id style, e.g. "gw1", "co1"),
// the owning tenant id, and audit timestamps. Every business record is
// associated with exactly one Tenant via TenantID (Req 3.2).
type TenantBase struct {
	ID        string `gorm:"primaryKey" json:"id"`
	TenantID  string `gorm:"index;not null" json:"tenantId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
