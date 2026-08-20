package api

import (
	"encoding/json"
	"strings"
	"testing"

	"pgcs/backend/internal/model"
	"pgcs/backend/internal/service"
)

// sampleMerchant builds a merchant carrying every secret-bearing field so the
// redaction tests can assert each one is withheld or preserved.
func sampleMerchant() model.Merchant {
	return model.Merchant{
		TenantBase: model.TenantBase{ID: "m1", TenantID: "t1"},
		Name:       "Acme",
		Banks: []model.MerchantBank{{
			ID:            "b1",
			LoginPassword: "login-secret",
			TxnPassword:   "txn-secret",
			Mpin:          "1234",
			AtmCards: []model.AtmCard{{
				ID:     "c1",
				Cvv:    "999",
				AtmPin: "4321",
			}},
		}},
		PaymentGateways: []model.MerchantGatewayCredential{{
			ID:       "g1",
			Password: "gw-secret",
		}},
	}
}

func TestMerchantSecretsVisible(t *testing.T) {
	m := sampleMerchant()
	cases := []struct {
		name string
		p    service.Principal
		want bool
	}{
		{"admin", service.Principal{Role: service.RoleAdmin}, true},
		{"superadmin", service.Principal{Role: service.RoleSuperAdmin}, true},
		{"owning merchant", service.Principal{Role: service.RoleMerchant, OwnerID: "m1"}, true},
		{"other merchant", service.Principal{Role: service.RoleMerchant, OwnerID: "m2"}, false},
		{"company", service.Principal{Role: service.RoleCompany, OwnerID: "co1"}, false},
		{"affiliate", service.Principal{Role: service.RoleAffiliate, OwnerID: "af1"}, false},
	}
	for _, tc := range cases {
		if got := merchantSecretsVisible(tc.p, &m); got != tc.want {
			t.Errorf("%s: merchantSecretsVisible = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestRedactMerchantSecretsWithholdsForNonOwner verifies every secret field is
// cleared (and therefore omitted from JSON via omitempty) for a principal that
// does not own the merchant (Req 8.7).
func TestRedactMerchantSecretsWithholdsForNonOwner(t *testing.T) {
	m := sampleMerchant()
	redactMerchantSecrets(service.Principal{Role: service.RoleMerchant, OwnerID: "other"}, &m)

	b := m.Banks[0]
	if b.LoginPassword != "" || b.TxnPassword != "" || b.Mpin != "" {
		t.Errorf("bank secrets not cleared: %+v", b)
	}
	if b.AtmCards[0].Cvv != "" || b.AtmCards[0].AtmPin != "" {
		t.Errorf("atm card secrets not cleared: %+v", b.AtmCards[0])
	}
	if m.PaymentGateways[0].Password != "" {
		t.Errorf("gateway credential secret not cleared: %+v", m.PaymentGateways[0])
	}

	// With the secrets cleared and tagged omitempty, none of the secret keys
	// should appear in the serialized body.
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, key := range []string{"loginPassword", "txnPassword", "mpin", "cvv", "atmPin", `"password"`} {
		if strings.Contains(string(out), key) {
			t.Errorf("serialized merchant still contains secret key %q: %s", key, out)
		}
	}
}

// TestRedactMerchantSecretsPreservesForOwner verifies secrets are left intact
// for an owning/admin principal (Req 8.7).
func TestRedactMerchantSecretsPreservesForOwner(t *testing.T) {
	m := sampleMerchant()
	redactMerchantSecrets(service.Principal{Role: service.RoleAdmin}, &m)

	if m.Banks[0].LoginPassword != "login-secret" || m.Banks[0].Mpin != "1234" {
		t.Errorf("admin should retain bank secrets, got %+v", m.Banks[0])
	}
	if m.Banks[0].AtmCards[0].Cvv != "999" {
		t.Errorf("admin should retain atm card secrets, got %+v", m.Banks[0].AtmCards[0])
	}
	if m.PaymentGateways[0].Password != "gw-secret" {
		t.Errorf("admin should retain gateway secret, got %+v", m.PaymentGateways[0])
	}
}
