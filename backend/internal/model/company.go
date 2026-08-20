package model

// Company mirrors the seed.js company shape. Its gateway assignments carry the
// per-gateway commission percentage and charge bearer used by the commission
// engine (calc.js gwAssign).
type Company struct {
	TenantBase
	Name             string `json:"name"`
	ContactPerson    string `json:"contactPerson"`
	ContactNumber    string `json:"contactNumber"`
	Whatsapp         string `json:"whatsapp"`
	Telegram         string `json:"telegram"`
	Email            string `json:"email"`
	AltContactPerson string `json:"altContactPerson"`
	AltContactNumber string `json:"altContactNumber"`
	Address          string `json:"address"`
	Status           string `json:"status"`
	// UserID is the login user id for this company's portal account. It is
	// stored on the entity for display and kept in sync with the owning
	// Account row. Password is write-only: accepted from create/update requests
	// to set/reset the portal account password, never persisted on the entity
	// (gorm:"-") and never serialized back (cleared before responses).
	UserID   string `json:"userId"`
	Password string `gorm:"-" json:"password,omitempty"`
	Gateways []CompanyGateway `gorm:"foreignKey:CompanyID" json:"gateways"`
}

// CompanyGateway is a company's assignment of a gateway, with the commission
// percentage and charge bearer (seed.js company.gateways entries).
type CompanyGateway struct {
	ID           string  `gorm:"primaryKey" json:"id"`
	TenantID     string  `gorm:"index" json:"tenantId"`
	CompanyID    string  `gorm:"index" json:"companyId"`
	GatewayID    string  `json:"gatewayId"`
	Commission   float64 `json:"commission"` // percent
	ChargeBearer string  `json:"chargeBearer"` // Admin | Company
}
