package model

// The methods below let the generic, tenant-scoped repository helpers
// (internal/repo) read a record's identity and stamp its owning tenant without
// reflection. Every primary business entity embeds TenantBase, so its pointer
// type automatically satisfies the repository's Entity constraint.

// GetID returns the record's primary key.
func (b *TenantBase) GetID() string { return b.ID }

// SetID assigns the record's primary key. The CRUD services call this to fill
// in a generated id when a create request omits one, so new records never
// collide on an empty primary key (seeded records keep their existing ids).
func (b *TenantBase) SetID(id string) { b.ID = id }

// GetTenantID returns the id of the tenant that owns the record.
func (b *TenantBase) GetTenantID() string { return b.TenantID }

// SetTenantID assigns the owning tenant. The repository calls this on create
// and update so the tenant is always taken from the authenticated principal,
// never from client-supplied input (Req 4.2).
func (b *TenantBase) SetTenantID(id string) { b.TenantID = id }
