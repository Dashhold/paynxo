package model

// Affiliate mirrors the seed.js affiliate shape. Its commission percentage and
// base drive beneficiary commission for affiliate-assigned merchants (calc.js).
type Affiliate struct {
	TenantBase
	Name           string  `json:"name"`
	Contact        string  `json:"contact"`
	AltContact     string  `json:"altContact"`
	Email          string  `json:"email"`
	CommissionPct  float64 `json:"commissionPct"`
	CommissionBase string  `json:"commissionBase"` // Transaction Amount | Settlement Amount
	Status         string  `json:"status"`
	// UserID / Password back the affiliate's portal login account (see Company
	// for the storage/serialization rules).
	UserID   string `json:"userId"`
	Password string `gorm:"-" json:"password,omitempty"`
}
