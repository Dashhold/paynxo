package model

// Merchant mirrors the seed.js merchant shape with its nested data owned by the
// parent merchant (Req 3.3). AffiliateID is a nullable pointer: nil for a
// direct merchant, set when the merchant is assigned to an affiliate (Req 3.5).
type Merchant struct {
	TenantBase
	Name           string  `json:"name"`
	Contact        string  `json:"contact"`
	AltContact     string  `json:"altContact"`
	Email          string  `json:"email"`
	CompanyID      string  `gorm:"index" json:"companyId"`
	AffiliateID    *string `gorm:"index" json:"affiliateId"` // null when direct merchant (Req 3.5)
	CommissionPct  float64 `json:"commissionPct"`
	CommissionBase string  `json:"commissionBase"` // Transaction Amount | Settlement Amount
	Status         string  `json:"status"`
	// UserID / Password back the merchant's portal login account (see Company
	// for the storage/serialization rules).
	UserID   string `json:"userId"`
	Password string `gorm:"-" json:"password,omitempty"`
	Banks           []MerchantBank              `gorm:"foreignKey:MerchantID" json:"banks"`
	PaymentGateways []MerchantGatewayCredential `gorm:"foreignKey:MerchantID" json:"paymentGateways"`
}

// MerchantBank is a bank account owned by a merchant. Preserves every seed.js
// bank field including mobileBanking, mobileLoginId, and mpin. It owns nested
// ATM cards and polymorphic custom fields.
type MerchantBank struct {
	ID             string `gorm:"primaryKey" json:"id"`
	TenantID       string `gorm:"index" json:"tenantId"`
	MerchantID     string `gorm:"index" json:"merchantId"`
	BankName       string `json:"bankName"`
	AccountName    string `json:"accountName"`
	AccountNumber  string `json:"accountNumber"`
	Ifsc           string `json:"ifsc"`
	NetbankingLink string `json:"netbankingLink"`
	Username       string `json:"username"`
	// Secret-bearing fields carry omitempty so the response projection can
	// withhold them from principals outside the owning tenant by clearing the
	// value, dropping the field from the JSON body entirely (Req 8.7).
	LoginPassword  string `json:"loginPassword,omitempty"`
	TxnPassword    string `json:"txnPassword,omitempty"`
	CustomerID     string `json:"customerId"`
	Mobile         string `json:"mobile"`
	Email          string `json:"email"`
	MobileBanking  string `json:"mobileBanking"`
	MobileLoginID  string `json:"mobileLoginId"`
	Mpin           string `json:"mpin,omitempty"`
	AtmCards []AtmCard     `gorm:"foreignKey:MerchantBankID" json:"atmCards"`
	Custom   []CustomField `gorm:"polymorphic:Owner;" json:"custom"`
}

// AtmCard is an ATM/debit card attached to a merchant bank account, preserving
// every seed.js card field.
type AtmCard struct {
	ID             string `gorm:"primaryKey" json:"id"`
	TenantID       string `gorm:"index" json:"tenantId"`
	MerchantBankID string `gorm:"index" json:"merchantBankId"`
	NameOnCard     string `json:"nameOnCard"`
	CardNumber     string `json:"cardNumber"`
	Expiry         string `json:"expiry"`
	Cvv            string `json:"cvv,omitempty"`    // secret; withheld outside owning tenant (Req 8.7)
	AtmPin         string `json:"atmPin,omitempty"` // secret; withheld outside owning tenant (Req 8.7)
}

// MerchantGatewayCredential is a merchant's login/credential for a payment
// gateway (seed.js merchant.paymentGateways entries). The seed "merchantId"
// field (the merchant's reference at the gateway) is stored as MerchantRef to
// avoid colliding with the MerchantID foreign key.
type MerchantGatewayCredential struct {
	ID          string `gorm:"primaryKey" json:"id"`
	TenantID    string `gorm:"index" json:"tenantId"`
	MerchantID  string `gorm:"index" json:"merchantId"`
	GatewayID   string `json:"gatewayId"`
	LoginLink   string `json:"loginLink"`
	MerchantRef string `json:"merchantRef"` // seed.js "merchantId" (gateway-side reference)
	Username    string `json:"username"`
	Password    string `json:"password,omitempty"` // secret; withheld outside owning tenant (Req 8.7)
	Mobile      string `json:"mobile"`
	Email       string `json:"email"`
	Custom []CustomField `gorm:"polymorphic:Owner;" json:"custom"`
}

// CustomField is a polymorphic key/value pair owned by either a MerchantBank or
// a MerchantGatewayCredential (seed.js "custom" arrays). GORM sets OwnerID and
// OwnerType from the polymorphic association.
type CustomField struct {
	ID        string `gorm:"primaryKey" json:"id"`
	TenantID  string `gorm:"index" json:"tenantId"`
	OwnerID   string `gorm:"index" json:"ownerId"`
	OwnerType string `json:"ownerType"`
	Label     string `json:"label"`
	Value     string `json:"value"`
}
