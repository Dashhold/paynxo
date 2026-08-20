package model

// Transaction mirrors the seed.js transaction shape. Date is kept as an ISO
// yyyy-mm-dd string, preserved from seed.
type Transaction struct {
	TenantBase
	CompanyID        string  `gorm:"index" json:"companyId"`
	MerchantID       string  `gorm:"index" json:"merchantId"`
	GatewayID        string  `gorm:"index" json:"gatewayId"`
	Date             string  `json:"date"` // ISO yyyy-mm-dd
	TxnAmount        float64 `json:"txnAmount"`
	SettlementAmount float64 `json:"settlementAmount"`
	TxnCharges       float64 `json:"txnCharges"`
	OtherCharges     float64 `json:"otherCharges"`
	Remarks          string  `json:"remarks"`
	// RecordedByRole records which role entered the transaction ("Admin",
	// "SuperAdmin", or "Company"), so a company-entered record is
	// distinguishable from one the admin entered.
	RecordedByRole string `json:"recordedByRole"`
}

// Payment-type values: whether money moved as physical Cash or through a Bank
// (any electronic/instrument transfer). Every payment record carries one, and a
// Bank payment additionally identifies the Bank it was routed through.
const (
	PaymentTypeCash = "Cash"
	PaymentTypeBank = "Bank"
)

// PaymentDetails is embedded by every payment record (settlements, affiliate
// commission payments, merchant commission payments). It captures HOW the money
// moved:
//
//   - PaymentType is the coarse channel: "Cash" or "Bank".
//   - BankID references the tenant's Bank the transfer was routed through. It is
//     required for a Bank payment and must be empty for a Cash payment.
//   - PaymentMode is the finer-grained instrument kept from the original schema
//     ("NEFT", "UPI", "Cheque", ...); for a Cash payment it is "Cash".
//   - RefNumber is the transfer/instrument reference; not applicable to Cash.
type PaymentDetails struct {
	PaymentType string `gorm:"index" json:"paymentType"` // Cash | Bank
	BankID      string `gorm:"index" json:"bankId"`
	PaymentMode string `json:"paymentMode"`
	RefNumber   string `json:"refNumber"`
}

// Settlement is a payment made to a company (seed.js settlements).
type Settlement struct {
	TenantBase
	CompanyID string  `gorm:"index" json:"companyId"`
	Date      string  `json:"date"`
	Amount    float64 `json:"amount"`
	PaymentDetails
	Remarks string `json:"remarks"`
	// RecordedByRole records which role entered the settlement ("Admin",
	// "SuperAdmin", or "Company"), so a company-entered record is
	// distinguishable from one the admin entered.
	RecordedByRole string `json:"recordedByRole"`
}

// AffiliatePayment is a commission payment made to an affiliate (seed.js
// affiliatePayments).
type AffiliatePayment struct {
	TenantBase
	AffiliateID string  `gorm:"index" json:"affiliateId"`
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	PaymentDetails
	Remarks string `json:"remarks"`
}

// MerchantPayment is a commission payment made to a merchant (seed.js
// merchantPayments).
type MerchantPayment struct {
	TenantBase
	MerchantID string  `gorm:"index" json:"merchantId"`
	Date       string  `json:"date"`
	Amount     float64 `json:"amount"`
	PaymentDetails
	Remarks string `json:"remarks"`
}
