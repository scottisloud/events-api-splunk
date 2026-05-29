package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.1password.io/eventsapi-splunk/utils"
)

func TestLoadTenantsLegacyConfig(t *testing.T) {
	dir := t.TempDir()
	confPath := filepath.Join(dir, "etc/apps/onepassword_events_api/local/events_reporting.conf")
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		t.Fatal(err)
	}
	content := `
[config]
limit = 50
startAt = 2020-01-01T00:00:00Z
signInCursorFile = "/etc/apps/onepassword_events_api/local/signin_cursor_store"
itemUsageCursorFile = "/etc/apps/onepassword_events_api/local/itemusage_cursor_store"
auditEventsCursorFile = "/etc/apps/onepassword_events_api/local/auditevents_cursor_store"
`
	if err := os.WriteFile(confPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	env, err := NewSplunkEnv(dir)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := env.LoadTenants()
	if err != nil {
		t.Fatal(err)
	}
	if len(tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d", len(tenants))
	}
	if tenants[0].TenantKey != utils.DefaultTenantKey {
		t.Fatalf("expected default key, got %q", tenants[0].TenantKey)
	}
	if tenants[0].TenantID != utils.DefaultTenantID {
		t.Fatalf("expected default id, got %q", tenants[0].TenantID)
	}
	if tenants[0].SecretName != utils.LegacySecretName {
		t.Fatalf("expected legacy secret, got %q", tenants[0].SecretName)
	}
}

func TestLoadTenantsMultiStanza(t *testing.T) {
	dir := t.TempDir()
	confPath := filepath.Join(dir, "etc/apps/onepassword_events_api/local/events_reporting.conf")
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		t.Fatal(err)
	}
	content := `
[tenant.acme]
tenantId = "acme-corp"
limit = 25
startAt = 2021-06-01T00:00:00Z
`
	if err := os.WriteFile(confPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	env, err := NewSplunkEnv(dir)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := env.LoadTenants()
	if err != nil {
		t.Fatal(err)
	}
	if len(tenants) != 1 {
		t.Fatalf("expected 1 tenant from [tenant.*], got %d", len(tenants))
	}
	if tenants[0].TenantID != "acme-corp" {
		t.Fatalf("got tenant id %q", tenants[0].TenantID)
	}
	if tenants[0].Config.Limit != 25 {
		t.Fatalf("expected limit 25, got %d", tenants[0].Config.Limit)
	}
	if tenants[0].Config.StartAt != time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected startAt %v", tenants[0].Config.StartAt)
	}
}

func TestLoadTenantsKeepsLegacyDefaultWithAdditionalTenants(t *testing.T) {
	dir := t.TempDir()
	confPath := filepath.Join(dir, "etc/apps/onepassword_events_api/local/events_reporting.conf")
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		t.Fatal(err)
	}
	content := `
[config]
limit = 100
signInCursorFile = "/etc/apps/onepassword_events_api/local/signin_cursor_store"

[tenant.acme]
tenantId = "acme-corp"
`
	if err := os.WriteFile(confPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	env, err := NewSplunkEnv(dir)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := env.LoadTenants()
	if err != nil {
		t.Fatal(err)
	}
	if len(tenants) != 2 {
		t.Fatalf("expected default + acme, got %d", len(tenants))
	}
}

func TestLoadTenantsSortedOrder(t *testing.T) {
	dir := t.TempDir()
	confPath := filepath.Join(dir, "etc/apps/onepassword_events_api/local/events_reporting.conf")
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		t.Fatal(err)
	}
	content := `
[tenant.zebra]
tenantId = "z"

[tenant.acme]
tenantId = "a"

[tenant.middle]
tenantId = "m"
`
	if err := os.WriteFile(confPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	env, err := NewSplunkEnv(dir)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := env.LoadTenants()
	if err != nil {
		t.Fatal(err)
	}
	if len(tenants) != 3 {
		t.Fatalf("expected 3 tenants, got %d", len(tenants))
	}
	if tenants[0].TenantKey != "acme" || tenants[1].TenantKey != "middle" || tenants[2].TenantKey != "zebra" {
		t.Fatalf("tenants not sorted by key: %#v", tenants)
	}
}

func TestLoadTenantsManyTenants(t *testing.T) {
	dir := t.TempDir()
	confPath := filepath.Join(dir, "etc/apps/onepassword_events_api/local/events_reporting.conf")
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		t.Fatal(err)
	}

	var content strings.Builder
	for i := 1; i <= 19; i++ {
		key := fmt.Sprintf("tenant_%02d", i)
		content.WriteString(fmt.Sprintf(`
[tenant.%s]
tenantId = "org-%02d"
`, key, i))
	}
	if err := os.WriteFile(confPath, []byte(content.String()), 0600); err != nil {
		t.Fatal(err)
	}

	env, err := NewSplunkEnv(dir)
	if err != nil {
		t.Fatal(err)
	}
	tenants, err := env.LoadTenants()
	if err != nil {
		t.Fatal(err)
	}
	if len(tenants) != 19 {
		t.Fatalf("expected 19 tenants, got %d", len(tenants))
	}
	for i, tenant := range tenants {
		wantKey := fmt.Sprintf("tenant_%02d", i+1)
		if tenant.TenantKey != wantKey {
			t.Fatalf("tenant %d: got key %q want %q", i, tenant.TenantKey, wantKey)
		}
	}
}
