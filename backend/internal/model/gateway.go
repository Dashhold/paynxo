package model

// Gateway is a payment gateway available within a tenant (seed.js gateways).
type Gateway struct {
	TenantBase
	Name   string `json:"name"`
	Status string `json:"status"` // Active | Inactive
}
