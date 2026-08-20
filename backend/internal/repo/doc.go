// Package repo provides the tenant-scoped GORM data-access layer. Every
// business-entity query is built from a tenant-scoped *gorm.DB so isolation
// cannot be bypassed. Implemented in task 3.1.
package repo
