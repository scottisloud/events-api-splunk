package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.1password.io/eventsapi-splunk/utils"
)

const appLocalRelPath = "etc/apps/onepassword_events_api/local"

// AppLocalDir returns the absolute path to the app's local configuration directory.
func AppLocalDir(splunkHome string) string {
	return filepath.Join(splunkHome, filepath.FromSlash(appLocalRelPath))
}

// ResolveCursorFile maps a configured cursor path to an absolute filesystem path
// under $SPLUNK_HOME/etc/apps/onepassword_events_api/local/.
func ResolveCursorFile(splunkHome, configPath string) (string, error) {
	p := strings.Trim(configPath, `"`)
	p = strings.TrimPrefix(p, "/")
	p = filepath.FromSlash(p)

	abs := filepath.Clean(filepath.Join(splunkHome, p))
	base := filepath.Clean(AppLocalDir(splunkHome))

	if abs != base && !strings.HasPrefix(abs, base+string(os.PathSeparator)) {
		return "", fmt.Errorf("cursor path %q is outside %q", configPath, base)
	}
	return abs, nil
}

// DeleteTenantCursorFiles removes per-tenant cursor files from the app local directory.
func DeleteTenantCursorFiles(splunkHome, tenantKey string) error {
	if err := utils.ValidateTenantKey(tenantKey); err != nil {
		return err
	}

	base := AppLocalDir(splunkHome)
	for _, name := range []string{"signin_cursor_store", "itemusage_cursor_store", "auditevents_cursor_store"} {
		filePath := filepath.Join(base, fmt.Sprintf("%s_%s", name, tenantKey))
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not remove cursor file %q: %w", filePath, err)
		}
	}
	return nil
}
