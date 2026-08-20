package crud

import (
	"encoding/json"
	"testing"

	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service/commission"
)

// TestTransactionWithBreakdownJSONShape verifies the read DTO promotes the
// transaction's own fields to the top level and nests the computed commission
// breakdown under "breakdown" (the {...transaction, breakdown: {...}} shape).
func TestTransactionWithBreakdownJSONShape(t *testing.T) {
	view := TransactionWithBreakdown{
		Transaction: model.Transaction{
			TenantBase: model.TenantBase{ID: "tx1", TenantID: "t1"},
			CompanyID:  "co1",
			MerchantID: "m1",
			GatewayID:  "gw1",
			TxnAmount:  1000,
		},
		Breakdown: commission.Breakdown{AdminCommission: 25, CompanyNetIncome: 975},
	}

	out, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Transaction fields are promoted to the top level.
	if got["id"] != "tx1" {
		t.Errorf("top-level id = %v, want tx1", got["id"])
	}
	if got["companyId"] != "co1" {
		t.Errorf("top-level companyId = %v, want co1", got["companyId"])
	}

	// The breakdown is nested under "breakdown".
	bd, ok := got["breakdown"].(map[string]any)
	if !ok {
		t.Fatalf("breakdown missing or wrong type: %v", got["breakdown"])
	}
	if bd["adminCommission"].(float64) != 25 {
		t.Errorf("breakdown.adminCommission = %v, want 25", bd["adminCommission"])
	}
	if bd["companyNetIncome"].(float64) != 975 {
		t.Errorf("breakdown.companyNetIncome = %v, want 975", bd["companyNetIncome"])
	}
}
