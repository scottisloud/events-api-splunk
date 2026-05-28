package config

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestNormalizeUnquotedTenantId(t *testing.T) {
	conf := `
[config]
limit = 100

[tenant.events_1password_com]
tenantId = mspc
enabled = true
`
	raw, err := decodeEventsReportingConf([]byte(conf))
	if err != nil {
		t.Fatal(err)
	}
	if raw.Tenants["events_1password_com"].TenantID != "mspc" {
		t.Fatalf("got %q", raw.Tenants["events_1password_com"].TenantID)
	}
}

func TestTenantIdMustBeQuotedInToml(t *testing.T) {
	type TC struct {
		TenantID string `toml:"tenantId"`
	}
	var v struct {
		Tenant map[string]TC `toml:"tenant"`
	}
	_, err := toml.Decode("[tenant.events_1password_com]\ntenantId = mspc", &v)
	if err == nil {
		t.Fatal("expected unquoted mspc to fail")
	}
	_, err = toml.Decode("[tenant.events_1password_com]\ntenantId = \"mspc\"", &v)
	if err != nil {
		t.Fatal(err)
	}
	if v.Tenant["events_1password_com"].TenantID != "mspc" {
		t.Fatalf("got %q", v.Tenant["events_1password_com"].TenantID)
	}
}
