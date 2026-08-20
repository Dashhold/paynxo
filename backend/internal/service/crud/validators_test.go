package crud

import (
	"errors"
	"testing"

	"pgcs/backend/internal/apierr"
	"pgcs/backend/internal/model"
)

// asValidation asserts that err is a 400 validation ServiceError naming the
// expected field, and returns it for further inspection.
func asValidation(t *testing.T, err error, field string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a validation error, got nil")
	}
	var se *apierr.ServiceError
	if !errors.As(err, &se) {
		t.Fatalf("error %v is not a *apierr.ServiceError", err)
	}
	if se.Kind != apierr.KindValidation {
		t.Fatalf("kind = %v, want validation", se.Kind)
	}
	if _, ok := se.Fields[field]; !ok {
		t.Fatalf("validation fields = %v, want it to identify %q", se.Fields, field)
	}
}

func TestValidateGateway(t *testing.T) {
	if err := ValidateGateway(&model.Gateway{Name: "Razorpay"}); err != nil {
		t.Errorf("valid gateway rejected: %v", err)
	}
	asValidation(t, ValidateGateway(&model.Gateway{Name: ""}), "name")
	asValidation(t, ValidateGateway(&model.Gateway{Name: "   "}), "name")
}

func TestValidateAffiliate(t *testing.T) {
	if err := ValidateAffiliate(&model.Affiliate{Name: "Acme"}); err != nil {
		t.Errorf("valid affiliate rejected: %v", err)
	}
	asValidation(t, ValidateAffiliate(&model.Affiliate{Name: ""}), "name")
}

func TestValidateCompany(t *testing.T) {
	if err := validateCompany(&model.Company{Name: "Globex"}); err != nil {
		t.Errorf("valid company rejected: %v", err)
	}
	asValidation(t, validateCompany(&model.Company{Name: ""}), "name")
}

func TestStampAssignmentsFillsTenantCompanyAndIDs(t *testing.T) {
	c := &model.Company{
		TenantBase: model.TenantBase{ID: "co1"},
		Gateways: []model.CompanyGateway{
			{GatewayID: "gw1"},                                   // missing id -> generated
			{ID: "cg-existing", GatewayID: "gw2"},                // keeps its id
		},
	}
	stampAssignments(c, "tenant-1")

	for i, g := range c.Gateways {
		if g.TenantID != "tenant-1" {
			t.Errorf("assignment %d tenant = %q, want tenant-1", i, g.TenantID)
		}
		if g.CompanyID != "co1" {
			t.Errorf("assignment %d company = %q, want co1", i, g.CompanyID)
		}
		if g.ID == "" {
			t.Errorf("assignment %d has empty id", i)
		}
	}
	if c.Gateways[1].ID != "cg-existing" {
		t.Errorf("existing assignment id = %q, want cg-existing", c.Gateways[1].ID)
	}
}

func TestGenIDUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenID()
		if id == "" {
			t.Fatal("GenID returned empty string")
		}
		if seen[id] {
			t.Fatalf("GenID returned a duplicate: %q", id)
		}
		seen[id] = true
	}
}
