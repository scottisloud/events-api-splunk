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

func TestValidateSecretNameForTenant(t *testing.T) {
	if err := ValidateSecretNameForTenant("events_api_token_acme", "acme"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSecretNameForTenant("events_api_token", "acme"); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestValidateTokenTenantKey(t *testing.T) {
	claims := &JWTClaims{Audience: []string{"events.1password.com"}}
	if err := ValidateTokenTenantKey(claims, "events_1password_com"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTokenTenantKey(claims, DefaultTenantKey); err != nil {
		t.Fatal("default tenant should skip audience binding")
	}
	if err := ValidateTokenTenantKey(claims, "wrong_key"); err == nil {
		t.Fatal("expected audience mismatch error")
	}
}
