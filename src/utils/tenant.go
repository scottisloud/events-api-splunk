package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

const (
	DefaultTenantKey = "default"
	DefaultTenantID  = "default"
	SecretRealm      = "events_reporting_realm"
	LegacySecretName = "events_api_token"
)

var tenantIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// TenantKeyFromAudience derives a stable filesystem-safe key from the JWT audience host.
func TenantKeyFromAudience(audience string) (string, error) {
	if audience == "" {
		return "", fmt.Errorf("audience is empty")
	}
	if audience == AudienceDEPRECATED {
		return "", fmt.Errorf("deprecated audience")
	}

	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, audience)

	slug = strings.Trim(slug, "_")
	for strings.Contains(slug, "__") {
		slug = strings.ReplaceAll(slug, "__", "_")
	}
	if slug == "" {
		hash := sha256.Sum256([]byte(audience))
		return "t_" + hex.EncodeToString(hash[:8]), nil
	}
	if len(slug) > 48 {
		hash := sha256.Sum256([]byte(audience))
		return slug[:32] + "_" + hex.EncodeToString(hash[:4]), nil
	}
	return slug, nil
}

// ValidateTenantID checks the admin-facing tenant label used in Splunk events.
func ValidateTenantID(tenantID string) error {
	if !tenantIDPattern.MatchString(tenantID) {
		return fmt.Errorf("tenant_id must match %s", tenantIDPattern.String())
	}
	return nil
}

// SecretNameForTenant returns the Splunk storage password name for a tenant key.
func SecretNameForTenant(tenantKey string) (string, error) {
	if tenantKey == DefaultTenantKey {
		return LegacySecretName, nil
	}
	if err := ValidateTenantKey(tenantKey); err != nil {
		return "", err
	}
	return LegacySecretName + "_" + tenantKey, nil
}

// ValidateTenantKey ensures internal keys are safe for stanza names and file paths.
func ValidateTenantKey(tenantKey string) error {
	if tenantKey == DefaultTenantKey {
		return nil
	}
	if !tenantIDPattern.MatchString(tenantKey) {
		return fmt.Errorf("tenant_key must match %s", tenantIDPattern.String())
	}
	return nil
}

// TenantIDFromAudience creates a default human-readable tenant_id from aud when none is provided.
func TenantIDFromAudience(audience string) (string, error) {
	id, err := TenantKeyFromAudience(audience)
	if err != nil {
		return "", err
	}
	return id, nil
}
