package config

import (
	"bytes"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
)

// Splunk writes events_reporting.conf without quoting some string values.
// BurntSushi TOML requires quotes for values like tenantId = mspc.
var unquotedTenantID = regexp.MustCompile(`(?m)^(\s*tenantId\s*=\s*)([a-zA-Z0-9][a-zA-Z0-9_-]*)\s*$`)
var unquotedPendingKeys = regexp.MustCompile(`(?m)^(\s*pendingKeys\s*=\s*)([a-zA-Z0-9][a-zA-Z0-9_,-]*)\s*$`)

func normalizeEventsReportingConf(data []byte) []byte {
	data = unquotedTenantID.ReplaceAll(data, []byte(`${1}"${2}"`))
	return unquotedPendingKeys.ReplaceAll(data, []byte(`${1}"${2}"`))
}

func decodeEventsReportingConf(data []byte) (rawConfigFile, error) {
	raw := rawConfigFile{}
	normalized := normalizeEventsReportingConf(data)
	_, err := toml.Decode(string(normalized), &raw)
	return raw, err
}

func readAndDecodeConfig(path string) (rawConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return rawConfigFile{}, err
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
	return decodeEventsReportingConf(data)
}
