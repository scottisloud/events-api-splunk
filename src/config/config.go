package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/dimchansky/utfbom"
	"go.1password.io/eventsapi-splunk/utils"
)

type Config struct {
	Url                   string
	AuthToken             string
	TenantID              string `toml:"tenantId"`
	StartAt               time.Time
	Limit                 int
	SignInCursorFile      string
	ItemUsageCursorFile   string
	AuditEventsCursorFile string
}

// TenantConfig holds per-tenant settings under [tenant.<tenant_key>] stanzas.
type TenantConfig struct {
	TenantID              string `toml:"tenantId"`
	Enabled               *bool  `toml:"enabled"`
	Url                   string
	SecretName            string `toml:"secretName"`
	StartAt               time.Time
	Limit                 int
	SignInCursorFile      string
	ItemUsageCursorFile   string
	AuditEventsCursorFile string
}

// TenantRuntime is a resolved tenant used during ingestion.
type TenantRuntime struct {
	TenantKey  string
	TenantID   string
	SecretName string
	Enabled    bool
	Config     Config
}

type SplunkEnv struct {
	Home       string
	ConfigPath string
	Config     Config
	Tenants    map[string]TenantConfig
}

type rawConfigFile struct {
	Config  Config                   `toml:"config"`
	Tenants map[string]TenantConfig  `toml:"tenant"`
}

const (
	defaultSignInCursor     = `"/etc/apps/onepassword_events_api/local/signin_cursor_store"`
	defaultItemUsageCursor  = `"/etc/apps/onepassword_events_api/local/itemusage_cursor_store"`
	defaultAuditEventsCursor = `"/etc/apps/onepassword_events_api/local/auditevents_cursor_store"`
)

// NewSplunkEnv loads events_reporting.conf including legacy [config] and [tenant.*] stanzas.
func NewSplunkEnv(splunkHome string) (*SplunkEnv, error) {
	log.Println("New Config")

	sc := SplunkEnv{
		Home:       splunkHome,
		ConfigPath: path.Join(splunkHome, "/etc/apps/onepassword_events_api/local/events_reporting.conf"),
		Tenants:    map[string]TenantConfig{},
	}

	configFile, err := os.Open(sc.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not open config file: %w", err)
	}
	defer configFile.Close()

	br := utfbom.SkipOnly(configFile)

	raw := rawConfigFile{}
	if _, err := toml.DecodeReader(br, &raw); err != nil {
		return nil, fmt.Errorf("could not decode toml config file: %w", err)
	}

	sc.Config = raw.Config
	if raw.Tenants != nil {
		sc.Tenants = raw.Tenants
	}

	return &sc, nil
}

// LoadTenants returns enabled tenants from [tenant.*] stanzas plus the legacy [config] default tenant when present.
func (e *SplunkEnv) LoadTenants() ([]TenantRuntime, error) {
	var runtimes []TenantRuntime

	if len(e.Tenants) > 0 {
		for tenantKey, tc := range e.Tenants {
			if err := utils.ValidateTenantKey(tenantKey); err != nil {
				return nil, fmt.Errorf("invalid tenant_key %q: %w", tenantKey, err)
			}
			enabled := true
			if tc.Enabled != nil {
				enabled = *tc.Enabled
			}
			if !enabled {
				continue
			}
			rt, err := e.buildTenantRuntime(tenantKey, tc, false)
			if err != nil {
				return nil, err
			}
			runtimes = append(runtimes, rt)
		}
	}

	if _, hasDefault := e.Tenants[utils.DefaultTenantKey]; !hasDefault && e.hasLegacyConfig() {
		rt, err := e.buildTenantRuntime(utils.DefaultTenantKey, TenantConfig{
			TenantID:              e.Config.TenantID,
			Limit:                 e.Config.Limit,
			StartAt:               e.Config.StartAt,
			Url:                   e.Config.Url,
			SignInCursorFile:      e.Config.SignInCursorFile,
			ItemUsageCursorFile:   e.Config.ItemUsageCursorFile,
			AuditEventsCursorFile: e.Config.AuditEventsCursorFile,
		}, true)
		if err != nil {
			return nil, err
		}
		runtimes = append(runtimes, rt)
	}

	if len(runtimes) == 0 {
		if !e.hasLegacyConfig() {
			return nil, fmt.Errorf("no tenants in configuration")
		}
		rt, err := e.buildTenantRuntime(utils.DefaultTenantKey, TenantConfig{
			TenantID:              e.Config.TenantID,
			Limit:                 e.Config.Limit,
			StartAt:               e.Config.StartAt,
			Url:                   e.Config.Url,
			SignInCursorFile:      e.Config.SignInCursorFile,
			ItemUsageCursorFile:   e.Config.ItemUsageCursorFile,
			AuditEventsCursorFile: e.Config.AuditEventsCursorFile,
		}, true)
		if err != nil {
			return nil, err
		}
		return []TenantRuntime{rt}, nil
	}

	return runtimes, nil
}

func (e *SplunkEnv) hasLegacyConfig() bool {
	return e.Config.Limit > 0 ||
		!e.Config.StartAt.IsZero() ||
		e.Config.SignInCursorFile != "" ||
		e.Config.AuthToken != ""
}

func (e *SplunkEnv) buildTenantRuntime(tenantKey string, tc TenantConfig, legacy bool) (TenantRuntime, error) {
	tenantID := tc.TenantID
	if tenantID == "" {
		tenantID = utils.DefaultTenantID
	}
	if err := utils.ValidateTenantID(tenantID); err != nil {
		return TenantRuntime{}, fmt.Errorf("tenant %q: %w", tenantKey, err)
	}

	secretName := tc.SecretName
	if secretName == "" {
		var err error
		secretName, err = utils.SecretNameForTenant(tenantKey)
		if err != nil {
			return TenantRuntime{}, err
		}
	}

	cfg := Config{
		Url:                   firstNonEmpty(tc.Url, e.Config.Url),
		Limit:                 firstPositiveInt(tc.Limit, e.Config.Limit, 100),
		StartAt:               firstNonZeroTime(tc.StartAt, e.Config.StartAt),
		SignInCursorFile:      cursorPath(tenantKey, tc.SignInCursorFile, e.Config.SignInCursorFile, legacy, "signin_cursor_store"),
		ItemUsageCursorFile:   cursorPath(tenantKey, tc.ItemUsageCursorFile, e.Config.ItemUsageCursorFile, legacy, "itemusage_cursor_store"),
		AuditEventsCursorFile: cursorPath(tenantKey, tc.AuditEventsCursorFile, e.Config.AuditEventsCursorFile, legacy, "auditevents_cursor_store"),
		TenantID:              tenantID,
	}

	if legacy && e.Config.AuthToken != "" {
		cfg.AuthToken = e.Config.AuthToken
	}

	return TenantRuntime{
		TenantKey:  tenantKey,
		TenantID:   tenantID,
		SecretName: secretName,
		Enabled:    true,
		Config:     cfg,
	}, nil
}

func cursorPath(tenantKey, tenantValue, legacyValue string, legacy bool, baseName string) string {
	if tenantValue != "" {
		return tenantValue
	}
	if legacyValue != "" && legacy {
		return legacyValue
	}
	if legacy || tenantKey == utils.DefaultTenantKey {
		switch baseName {
		case "signin_cursor_store":
			return defaultSignInCursor
		case "itemusage_cursor_store":
			return defaultItemUsageCursor
		default:
			return defaultAuditEventsCursor
		}
	}
	return fmt.Sprintf(`"/etc/apps/onepassword_events_api/local/%s_%s"`, baseName, tenantKey)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func firstPositiveInt(values ...int) int {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 100
}

func firstNonZeroTime(values ...time.Time) time.Time {
	for _, v := range values {
		if !v.IsZero() {
			return v
		}
	}
	return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
}

func (e *SplunkEnv) UpdateConfig(newConfig Config) error {
	configTemp := e.ConfigPath + "_" + strconv.Itoa(os.Getpid())
	configFile, err := os.Create(configTemp)
	if err != nil {
		return fmt.Errorf("could not create config file: %w", err)
	}
	defer configFile.Close()

	raw := rawConfigFile{
		Config:  newConfig,
		Tenants: e.Tenants,
	}
	encoder := toml.NewEncoder(configFile)
	if err := encoder.Encode(raw); err != nil {
		return fmt.Errorf("could not write toml to file: %w", err)
	}

	if err := os.Rename(configTemp, e.ConfigPath); err != nil {
		return fmt.Errorf("could not overwrite previous config: %w", err)
	}
	e.Config = newConfig
	return nil
}
