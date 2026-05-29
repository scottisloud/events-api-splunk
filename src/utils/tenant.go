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

// NormalizeTenantID strips optional surrounding quotes from config values.
func NormalizeTenantID(value string) string {
	normalized := strings.TrimSpace(value)
	if len(normalized) >= 2 && strings.HasPrefix(normalized, `"`) && strings.HasSuffix(normalized, `"`) {
		return normalized[1 : len(normalized)-1]
	}
	return normalized
}

// ValidateTenantID checks the admin-facing tenant label used in Splunk events.
func ValidateTenantID(tenantID string) error {
	tenantID = NormalizeTenantID(tenantID)
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

// ValidateSecretNameForTenant ensures a configured secret name matches the tenant key.
func ValidateSecretNameForTenant(secretName, tenantKey string) error {
	expected, err := SecretNameForTenant(tenantKey)
	if err != nil {
		return err
	}
	if secretName != expected {
		return fmt.Errorf("secretName %q does not match tenant_key %q (expected %q)", secretName, tenantKey, expected)
	}
	return nil
}

// ValidateTokenAudience ensures the JWT has a usable Events API audience.
// Multiple tenants may share the same endpoint host; tenant_key is the admin label.
func ValidateTokenAudience(claims *JWTClaims) error {
	if len(claims.Audience) == 0 {
		return fmt.Errorf("token missing audience")
	}
	if claims.Audience[0] == AudienceDEPRECATED {
		return fmt.Errorf("deprecated audience")
	}
	_, err := TenantKeyFromAudience(claims.Audience[0])
	if err != nil {
		return fmt.Errorf("invalid token audience: %w", err)
	}
	return nil
}
