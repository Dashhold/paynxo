package crud

import (
	"strings"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
)

// ValidateGateway enforces a gateway's required fields, returning a 400
// validation error identifying the offending field (Req 8.4). A gateway must
// have a name.
func ValidateGateway(g *model.Gateway) error {
	if strings.TrimSpace(g.Name) == "" {
		return apierr.ValidationField("name", "gateway name is required")
	}
	return nil
}

// ValidateBank enforces a bank's required fields, returning a 400 validation
// error identifying the offending field (Req 8.4). A bank must have a name and
// a code; the SWIFT code is optional.
func ValidateBank(b *model.Bank) error {
	if strings.TrimSpace(b.Name) == "" {
		return apierr.ValidationField("name", "bank name is required")
	}
	if strings.TrimSpace(b.Code) == "" {
		return apierr.ValidationField("code", "bank code is required")
	}
	return nil
}

// ValidateAffiliate enforces an affiliate's required fields, returning a 400
// validation error identifying the offending field (Req 8.4). An affiliate
// must have a name.
func ValidateAffiliate(a *model.Affiliate) error {
	if strings.TrimSpace(a.Name) == "" {
		return apierr.ValidationField("name", "affiliate name is required")
	}
	return nil
}

// NormalizePaymentDetails fills in and validates the Cash/Bank payment details
// shared by every payment record, returning a 400 validation error identifying
// the offending field (Req 8.4).
//
// Rules:
//   - PaymentType, when supplied, must be "Cash" or "Bank".
//   - A payment explicitly declared as "Bank" must reference a bank, so a client
//     that offers the Cash/Bank choice cannot record a bank transfer without
//     saying which bank it went through.
//   - A Cash payment never routes through a bank and carries no transfer
//     reference, so BankID/RefNumber are cleared rather than rejected.
//   - An omitted PaymentType is INFERRED from the legacy free-text PaymentMode
//     ("Cash" -> Cash, anything else -> Bank) so records written before the
//     field existed, and clients that only send a mode, keep working. An
//     inferred Bank payment does not require a bank, because the caller was
//     never given the chance to supply one.
//   - PaymentMode defaults to the payment type when the client omits it.
func NormalizePaymentDetails(d *model.PaymentDetails) error {
	d.PaymentType = strings.TrimSpace(d.PaymentType)
	d.BankID = strings.TrimSpace(d.BankID)
	d.PaymentMode = strings.TrimSpace(d.PaymentMode)
	d.RefNumber = strings.TrimSpace(d.RefNumber)

	// Whether the caller explicitly chose a type; drives bank enforcement below.
	declared := d.PaymentType != ""
	if !declared {
		if strings.EqualFold(d.PaymentMode, model.PaymentTypeCash) {
			d.PaymentType = model.PaymentTypeCash
		} else {
			d.PaymentType = model.PaymentTypeBank
		}
	}

	switch d.PaymentType {
	case model.PaymentTypeCash:
		d.BankID = ""
		d.RefNumber = ""
		if d.PaymentMode == "" {
			d.PaymentMode = model.PaymentTypeCash
		}
	case model.PaymentTypeBank:
		if declared && d.BankID == "" {
			return apierr.ValidationField("bankId", "bank is required for a bank payment")
		}
		if d.PaymentMode == "" {
			d.PaymentMode = "Bank Transfer"
		}
	default:
		return apierr.ValidationField("paymentType", `payment type must be "Cash" or "Bank"`)
	}
	return nil
}

// ValidateSettlement enforces a settlement's required fields, returning a 400
// validation error identifying the offending field (Req 8.4). A settlement is
// a payment to a company, so it must reference a company and carry a positive
// amount, plus valid Cash/Bank payment details.
func ValidateSettlement(s *model.Settlement) error {
	if strings.TrimSpace(s.CompanyID) == "" {
		return apierr.ValidationField("companyId", "settlement company is required")
	}
	if s.Amount <= 0 {
		return apierr.ValidationField("amount", "settlement amount must be greater than zero")
	}
	return NormalizePaymentDetails(&s.PaymentDetails)
}

// ValidateAffiliatePayment enforces an affiliate payment's required fields,
// returning a 400 validation error identifying the offending field (Req 8.4).
// It must reference an affiliate and carry a positive amount.
func ValidateAffiliatePayment(a *model.AffiliatePayment) error {
	if strings.TrimSpace(a.AffiliateID) == "" {
		return apierr.ValidationField("affiliateId", "affiliate payment affiliate is required")
	}
	if a.Amount <= 0 {
		return apierr.ValidationField("amount", "affiliate payment amount must be greater than zero")
	}
	return NormalizePaymentDetails(&a.PaymentDetails)
}

// ValidateMerchantPayment enforces a merchant payment's required fields,
// returning a 400 validation error identifying the offending field (Req 8.4).
// It must reference a merchant and carry a positive amount.
func ValidateMerchantPayment(m *model.MerchantPayment) error {
	if strings.TrimSpace(m.MerchantID) == "" {
		return apierr.ValidationField("merchantId", "merchant payment merchant is required")
	}
	if m.Amount <= 0 {
		return apierr.ValidationField("amount", "merchant payment amount must be greater than zero")
	}
	return NormalizePaymentDetails(&m.PaymentDetails)
}

// ValidateTransaction enforces a transaction's required fields, returning a 400
// validation error identifying the offending field (Req 8.4). A transaction
// must reference the company and merchant it belongs to; the gateway is
// optional (the commission engine defaults to a 0% gateway commission when no
// assignment matches).
func ValidateTransaction(t *model.Transaction) error {
	if strings.TrimSpace(t.CompanyID) == "" {
		return apierr.ValidationField("companyId", "transaction company is required")
	}
	if strings.TrimSpace(t.MerchantID) == "" {
		return apierr.ValidationField("merchantId", "transaction merchant is required")
	}
	return nil
}
