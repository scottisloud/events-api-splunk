package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCursorFileUnderAppLocal(t *testing.T) {
	home := "/opt/splunk"
	got, err := ResolveCursorFile(home, `"/etc/apps/onepassword_events_api/local/signin_cursor_store_acme"`)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "etc/apps/onepassword_events_api/local/signin_cursor_store_acme")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveCursorFileRejectsTraversal(t *testing.T) {
	home := "/opt/splunk"
	_, err := ResolveCursorFile(home, `"/etc/apps/onepassword_events_api/local/../../passwd"`)
	if err == nil {
		t.Fatal("expected error for path outside app local directory")
	}
}

func TestDeleteTenantCursorFiles(t *testing.T) {
	dir := t.TempDir()
	localDir := filepath.Join(dir, "etc/apps/onepassword_events_api/local")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		t.Fatal(err)
	}
	cursor := filepath.Join(localDir, "signin_cursor_store_acme")
	if err := os.WriteFile(cursor, []byte("cursor"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := DeleteTenantCursorFiles(dir, "acme"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cursor); !os.IsNotExist(err) {
		t.Fatalf("expected cursor file to be removed, stat err: %v", err)
	}
}
