package model

// Bank is a bank available within a tenant, used when configuring merchant
// settlement destinations. It mirrors the lightweight reference-entity shape of
// Gateway: a tenant-scoped record with a display name plus identifying codes.
type Bank struct {
	TenantBase
	Name      string `json:"name"`
	Code      string `json:"code"`
	SwiftCode string `json:"swiftCode"`
}
