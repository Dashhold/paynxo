package commission

import (
	"testing"

	"pgcs/backend/internal/model"
)

func ptr(s string) *string { return &s }

// company builds a Company with a single gateway assignment for gwID.
func company(gwID string, pct float64, bearer string) *model.Company {
	c := &model.Company{}
	c.Gateways = []model.CompanyGateway{
		{GatewayID: gwID, Commission: pct, ChargeBearer: bearer},
	}
	return c
}

// TestCalc_DirectMerchant_AdminBearer covers the common case: a direct merchant
// (no affiliate), admin bears the charges, settlement-amount base.
func TestCalc_DirectMerchant_AdminBearer(t *testing.T) {
	txn := model.Transaction{
		GatewayID:        "gw1",
		TxnAmount:        10000,
		SettlementAmount: 9500,
		TxnCharges:       50,
		OtherCharges:     25,
	}
	ctx := TxnContext{
		Company: company("gw1", 2, ChargeBearerAdmin),
		Merchant: &model.Merchant{
			CommissionPct:  1,
			CommissionBase: BaseSettlementAmount,
		},
	}

	b := Calc(txn, ctx)

	// adminCommission = 10000 * 2 / 100 = 200
	if b.AdminCommission != 200 {
		t.Errorf("AdminCommission = %v, want 200", b.AdminCommission)
	}
	if b.Beneficiary != BeneficiaryMerchant {
		t.Errorf("Beneficiary = %q, want Merchant", b.Beneficiary)
	}
	// beneficiaryCommission = 9500 * 1 / 100 = 95
	if b.BeneficiaryCommission != 95 {
		t.Errorf("BeneficiaryCommission = %v, want 95", b.BeneficiaryCommission)
	}
	// companyNetIncome = 9500 - 200 = 9300 (Admin bearer, no company deduction)
	if b.CompanyNetIncome != 9300 {
		t.Errorf("CompanyNetIncome = %v, want 9300", b.CompanyNetIncome)
	}
	if b.CompanyChargesDeducted != 0 {
		t.Errorf("CompanyChargesDeducted = %v, want 0", b.CompanyChargesDeducted)
	}
	// adminNetCommission = 200 - 95 - (50 + 25) = 30
	if b.AdminChargesDeducted != 75 {
		t.Errorf("AdminChargesDeducted = %v, want 75", b.AdminChargesDeducted)
	}
	if b.AdminNetCommission != 30 {
		t.Errorf("AdminNetCommission = %v, want 30", b.AdminNetCommission)
	}
	// ChargesDeducted alias mirrors CompanyChargesDeducted.
	if b.ChargesDeducted != b.CompanyChargesDeducted {
		t.Errorf("ChargesDeducted alias = %v, want %v", b.ChargesDeducted, b.CompanyChargesDeducted)
	}
}

// TestCalc_CompanyBearer_TransactionBase covers the company charge bearer and a
// Transaction Amount commission base.
func TestCalc_CompanyBearer_TransactionBase(t *testing.T) {
	txn := model.Transaction{
		GatewayID:        "gw1",
		TxnAmount:        20000,
		SettlementAmount: 19000,
		TxnCharges:       100,
		OtherCharges:     50,
	}
	ctx := TxnContext{
		Company: company("gw1", 3, ChargeBearerCompany),
		Merchant: &model.Merchant{
			CommissionPct:  2,
			CommissionBase: BaseTransactionAmount,
		},
	}

	b := Calc(txn, ctx)

	// adminCommission = 20000 * 3 / 100 = 600
	if b.AdminCommission != 600 {
		t.Errorf("AdminCommission = %v, want 600", b.AdminCommission)
	}
	// base = txnAmount -> beneficiaryCommission = 20000 * 2 / 100 = 400
	if b.BeneficiaryCommission != 400 {
		t.Errorf("BeneficiaryCommission = %v, want 400", b.BeneficiaryCommission)
	}
	// companyNetIncome = 19000 - 600 - (100 + 50) = 18250
	if b.CompanyChargesDeducted != 150 {
		t.Errorf("CompanyChargesDeducted = %v, want 150", b.CompanyChargesDeducted)
	}
	if b.CompanyNetIncome != 18250 {
		t.Errorf("CompanyNetIncome = %v, want 18250", b.CompanyNetIncome)
	}
	// adminNetCommission = 600 - 400 = 200 (Company bearer, no admin deduction)
	if b.AdminChargesDeducted != 0 {
		t.Errorf("AdminChargesDeducted = %v, want 0", b.AdminChargesDeducted)
	}
	if b.AdminNetCommission != 200 {
		t.Errorf("AdminNetCommission = %v, want 200", b.AdminNetCommission)
	}
}

// TestCalc_AffiliateBeneficiary covers an affiliate-assigned merchant: the
// affiliate is the beneficiary and its percentage/base sets the TOTAL commission
// the admin pays out, of which the merchant's own share is carved out.
func TestCalc_AffiliateBeneficiary(t *testing.T) {
	txn := model.Transaction{
		GatewayID:        "gw1",
		TxnAmount:        10000,
		SettlementAmount: 9000,
	}
	ctx := TxnContext{
		Company: company("gw1", 2, ChargeBearerAdmin),
		Merchant: &model.Merchant{
			AffiliateID:    ptr("aff1"),
			CommissionPct:  0.2,
			CommissionBase: BaseSettlementAmount,
		},
		Affiliate: &model.Affiliate{
			CommissionPct:  1,
			CommissionBase: BaseSettlementAmount,
		},
	}

	b := Calc(txn, ctx)

	if b.Beneficiary != BeneficiaryAffiliate {
		t.Errorf("Beneficiary = %q, want Affiliate", b.Beneficiary)
	}
	if b.BeneficiaryPct != 1 {
		t.Errorf("BeneficiaryPct = %v, want 1 (affiliate)", b.BeneficiaryPct)
	}
	// base = settlement -> 9000 * 1 / 100 = 90 total paid out by the admin.
	if b.BeneficiaryCommission != 90 {
		t.Errorf("BeneficiaryCommission = %v, want 90", b.BeneficiaryCommission)
	}
	if b.AffiliateGrossCommission != 90 {
		t.Errorf("AffiliateGrossCommission = %v, want 90", b.AffiliateGrossCommission)
	}
	// Merchant carve-out: 9000 * 0.2 / 100 = 18.
	if b.MerchantCommission != 18 {
		t.Errorf("MerchantCommission = %v, want 18", b.MerchantCommission)
	}
	// Affiliate keeps 90 - 18 = 72.
	if b.AffiliateCommission != 72 {
		t.Errorf("AffiliateCommission = %v, want 72", b.AffiliateCommission)
	}
	// The split must never change what the admin pays out in total.
	if b.MerchantCommission+b.AffiliateCommission != b.BeneficiaryCommission {
		t.Errorf("merchant %v + affiliate %v != total %v",
			b.MerchantCommission, b.AffiliateCommission, b.BeneficiaryCommission)
	}
}

// TestCalc_MerchantCarveOutCappedAtAffiliateGross verifies that a merchant
// percentage larger than the affiliate's cut is capped at that cut, so the
// affiliate's net never goes negative and the admin's payout stays the gross.
func TestCalc_MerchantCarveOutCappedAtAffiliateGross(t *testing.T) {
	txn := model.Transaction{GatewayID: "gw1", TxnAmount: 10000, SettlementAmount: 9000}
	ctx := TxnContext{
		Company: company("gw1", 2, ChargeBearerAdmin),
		Merchant: &model.Merchant{
			AffiliateID:    ptr("aff1"),
			CommissionPct:  5, // 450 — far more than the affiliate's 90
			CommissionBase: BaseSettlementAmount,
		},
		Affiliate: &model.Affiliate{CommissionPct: 1, CommissionBase: BaseSettlementAmount},
	}

	b := Calc(txn, ctx)

	if b.MerchantCommission != 90 {
		t.Errorf("MerchantCommission = %v, want 90 (capped at affiliate gross)", b.MerchantCommission)
	}
	if b.AffiliateCommission != 0 {
		t.Errorf("AffiliateCommission = %v, want 0", b.AffiliateCommission)
	}
	if b.BeneficiaryCommission != 90 {
		t.Errorf("BeneficiaryCommission = %v, want 90 (admin payout unchanged)", b.BeneficiaryCommission)
	}
}

// TestCalc_DirectMerchantCommissionEqualsBeneficiary verifies that for a direct
// merchant the merchant commission IS the whole beneficiary commission and no
// affiliate figures are produced.
func TestCalc_DirectMerchantCommissionEqualsBeneficiary(t *testing.T) {
	txn := model.Transaction{GatewayID: "gw1", TxnAmount: 10000, SettlementAmount: 9000}
	ctx := TxnContext{
		Company:  company("gw1", 2, ChargeBearerAdmin),
		Merchant: &model.Merchant{CommissionPct: 1, CommissionBase: BaseSettlementAmount},
	}

	b := Calc(txn, ctx)

	if b.MerchantCommission != 90 {
		t.Errorf("MerchantCommission = %v, want 90", b.MerchantCommission)
	}
	if b.MerchantCommission != b.BeneficiaryCommission {
		t.Errorf("MerchantCommission %v != BeneficiaryCommission %v",
			b.MerchantCommission, b.BeneficiaryCommission)
	}
	if b.AffiliateGrossCommission != 0 || b.AffiliateCommission != 0 {
		t.Errorf("affiliate figures should be 0 for a direct merchant, got gross %v net %v",
			b.AffiliateGrossCommission, b.AffiliateCommission)
	}
}

// TestCalc_AffiliateIDSetButNoAffiliate falls back to the merchant when the
// affiliate record is not supplied, mirroring calc.js (merchant?.affiliateId
// && affiliate).
func TestCalc_AffiliateIDSetButNoAffiliate(t *testing.T) {
	txn := model.Transaction{GatewayID: "gw1", TxnAmount: 10000, SettlementAmount: 9000}
	ctx := TxnContext{
		Company: company("gw1", 2, ChargeBearerAdmin),
		Merchant: &model.Merchant{
			AffiliateID:    ptr("aff1"),
			CommissionPct:  3,
			CommissionBase: BaseSettlementAmount,
		},
		Affiliate: nil,
	}

	b := Calc(txn, ctx)

	if b.Beneficiary != BeneficiaryMerchant {
		t.Errorf("Beneficiary = %q, want Merchant (affiliate missing)", b.Beneficiary)
	}
	// 9000 * 3 / 100 = 270
	if b.BeneficiaryCommission != 270 {
		t.Errorf("BeneficiaryCommission = %v, want 270", b.BeneficiaryCommission)
	}
}

// TestCalc_NilCompany_Defaults verifies graceful handling of a nil company: 0%
// gateway commission and an Admin charge bearer.
func TestCalc_NilCompany_Defaults(t *testing.T) {
	txn := model.Transaction{
		GatewayID:        "gw1",
		TxnAmount:        10000,
		SettlementAmount: 9000,
		TxnCharges:       40,
		OtherCharges:     10,
	}
	ctx := TxnContext{} // no company, no merchant, no affiliate

	b := Calc(txn, ctx)

	if b.GatewayCommissionPct != 0 {
		t.Errorf("GatewayCommissionPct = %v, want 0", b.GatewayCommissionPct)
	}
	if b.ChargeBearer != ChargeBearerAdmin {
		t.Errorf("ChargeBearer = %q, want Admin", b.ChargeBearer)
	}
	if b.AdminCommission != 0 {
		t.Errorf("AdminCommission = %v, want 0", b.AdminCommission)
	}
	// No merchant -> default beneficiary Merchant, 0%, Settlement Amount base.
	if b.Beneficiary != BeneficiaryMerchant || b.BeneficiaryPct != 0 {
		t.Errorf("beneficiary defaults wrong: %q pct=%v", b.Beneficiary, b.BeneficiaryPct)
	}
	if b.BeneficiaryBase != BaseSettlementAmount {
		t.Errorf("BeneficiaryBase = %q, want Settlement Amount", b.BeneficiaryBase)
	}
	// companyNetIncome = 9000 - 0 = 9000; adminNet = 0 - 0 - (40+10) = -50
	if b.CompanyNetIncome != 9000 {
		t.Errorf("CompanyNetIncome = %v, want 9000", b.CompanyNetIncome)
	}
	if b.AdminNetCommission != -50 {
		t.Errorf("AdminNetCommission = %v, want -50", b.AdminNetCommission)
	}
}

// TestCalc_GatewayNotAssigned defaults to 0%/Admin when the company has no
// assignment for the transaction's gateway.
func TestCalc_GatewayNotAssigned(t *testing.T) {
	txn := model.Transaction{GatewayID: "gwX", TxnAmount: 10000, SettlementAmount: 9000}
	ctx := TxnContext{
		Company:  company("gw1", 2, ChargeBearerCompany),
		Merchant: &model.Merchant{CommissionPct: 1, CommissionBase: BaseSettlementAmount},
	}

	b := Calc(txn, ctx)

	if b.GatewayCommissionPct != 0 {
		t.Errorf("GatewayCommissionPct = %v, want 0", b.GatewayCommissionPct)
	}
	if b.ChargeBearer != ChargeBearerAdmin {
		t.Errorf("ChargeBearer = %q, want Admin (no matching assignment)", b.ChargeBearer)
	}
}
