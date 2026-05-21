package utils

import "testing"

func TestTenantKeyFromAudience(t *testing.T) {
	key, err := TenantKeyFromAudience("events.1password.com")
	if err != nil {
		t.Fatal(err)
	}
	if key != "events_1password_com" {
		t.Fatalf("got %q", key)
	}
}

func TestValidateTenantID(t *testing.T) {
	if err := ValidateTenantID("acme-corp"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTenantID("../bad"); err == nil {
		t.Fatal("expected error for path-like tenant_id")
	}
}

func TestSecretNameForTenant(t *testing.T) {
	name, err := SecretNameForTenant(DefaultTenantKey)
	if err != nil || name != LegacySecretName {
		t.Fatalf("got %q err %v", name, err)
	}
	name, err = SecretNameForTenant("acme")
	if err != nil || name != "events_api_token_acme" {
		t.Fatalf("got %q err %v", name, err)
	}
}
