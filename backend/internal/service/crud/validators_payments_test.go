package crud

import (
	"errors"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
)

// fieldError asserts err is a 400 validation error naming the given field.
func fieldError(t *testing.T, err error, field string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error for field %q, got nil", field)
	}
	var se *apierr.ServiceError
	if !errors.As(err, &se) || se.Kind != apierr.KindValidation {
		t.Fatalf("expected validation error, got %T %v", err, err)
	}
	if _, ok := se.Fields[field]; !ok {
		t.Fatalf("expected validation error to name field %q, got fields %v", field, se.Fields)
	}
}

func TestValidateSettlement(t *testing.T) {
	if err := ValidateSettlement(&model.Settlement{CompanyID: "co1", Amount: 100}); err != nil {
		t.Errorf("valid settlement rejected: %v", err)
	}
	fieldError(t, ValidateSettlement(&model.Settlement{Amount: 100}), "companyId")
	fieldError(t, ValidateSettlement(&model.Settlement{CompanyID: "co1"}), "amount")
	fieldError(t, ValidateSettlement(&model.Settlement{CompanyID: "co1", Amount: -5}), "amount")
}

func TestValidateAffiliatePayment(t *testing.T) {
	if err := ValidateAffiliatePayment(&model.AffiliatePayment{AffiliateID: "af1", Amount: 50}); err != nil {
		t.Errorf("valid affiliate payment rejected: %v", err)
	}
	fieldError(t, ValidateAffiliatePayment(&model.AffiliatePayment{Amount: 50}), "affiliateId")
	fieldError(t, ValidateAffiliatePayment(&model.AffiliatePayment{AffiliateID: "af1"}), "amount")
}

func TestValidateMerchantPayment(t *testing.T) {
	if err := ValidateMerchantPayment(&model.MerchantPayment{MerchantID: "m1", Amount: 50}); err != nil {
		t.Errorf("valid merchant payment rejected: %v", err)
	}
	fieldError(t, ValidateMerchantPayment(&model.MerchantPayment{Amount: 50}), "merchantId")
	fieldError(t, ValidateMerchantPayment(&model.MerchantPayment{MerchantID: "m1"}), "amount")
}

// TestNormalizePaymentDetails_Bank verifies a caller that explicitly chooses
// "Bank" must name the bank, and that the mode defaults sensibly.
func TestNormalizePaymentDetails_Bank(t *testing.T) {
	fieldError(t, NormalizePaymentDetails(&model.PaymentDetails{
		PaymentType: model.PaymentTypeBank,
	}), "bankId")

	d := model.PaymentDetails{PaymentType: model.PaymentTypeBank, BankID: "bk1"}
	if err := NormalizePaymentDetails(&d); err != nil {
		t.Fatalf("valid bank payment rejected: %v", err)
	}
	if d.PaymentMode != "Bank Transfer" {
		t.Errorf("PaymentMode = %q, want the Bank Transfer default", d.PaymentMode)
	}
}

// TestNormalizePaymentDetails_CashClearsBankFields verifies a Cash payment drops
// any bank/reference the client may have sent rather than rejecting it.
func TestNormalizePaymentDetails_CashClearsBankFields(t *testing.T) {
	d := model.PaymentDetails{
		PaymentType: model.PaymentTypeCash,
		BankID:      "bk1",
		RefNumber:   "NEFT123",
	}
	if err := NormalizePaymentDetails(&d); err != nil {
		t.Fatalf("valid cash payment rejected: %v", err)
	}
	if d.BankID != "" {
		t.Errorf("BankID = %q, want it cleared for a cash payment", d.BankID)
	}
	if d.RefNumber != "" {
		t.Errorf("RefNumber = %q, want it cleared for a cash payment", d.RefNumber)
	}
	if d.PaymentMode != model.PaymentTypeCash {
		t.Errorf("PaymentMode = %q, want Cash", d.PaymentMode)
	}
}

// TestNormalizePaymentDetails_InfersLegacyType verifies a payload that predates
// the paymentType field still validates: the type is inferred from the mode and
// no bank is demanded, because the caller was never offered the choice.
func TestNormalizePaymentDetails_InfersLegacyType(t *testing.T) {
	cash := model.PaymentDetails{PaymentMode: "Cash"}
	if err := NormalizePaymentDetails(&cash); err != nil {
		t.Fatalf("legacy cash payload rejected: %v", err)
	}
	if cash.PaymentType != model.PaymentTypeCash {
		t.Errorf("PaymentType = %q, want Cash inferred from the mode", cash.PaymentType)
	}

	bank := model.PaymentDetails{PaymentMode: "NEFT"}
	if err := NormalizePaymentDetails(&bank); err != nil {
		t.Fatalf("legacy bank payload rejected: %v", err)
	}
	if bank.PaymentType != model.PaymentTypeBank {
		t.Errorf("PaymentType = %q, want Bank inferred from the mode", bank.PaymentType)
	}
}

// TestNormalizePaymentDetails_RejectsUnknownType keeps the enum closed.
func TestNormalizePaymentDetails_RejectsUnknownType(t *testing.T) {
	fieldError(t, NormalizePaymentDetails(&model.PaymentDetails{
		PaymentType: "Crypto",
	}), "paymentType")
}

func TestValidateTransaction(t *testing.T) {
	if err := ValidateTransaction(&model.Transaction{CompanyID: "co1", MerchantID: "m1"}); err != nil {
		t.Errorf("valid transaction rejected: %v", err)
	}
	fieldError(t, ValidateTransaction(&model.Transaction{MerchantID: "m1"}), "companyId")
	fieldError(t, ValidateTransaction(&model.Transaction{CompanyID: "co1"}), "merchantId")
}
