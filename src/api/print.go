package api

import (
	"encoding/json"
	"fmt"
)

// PrintJSONEvent writes one JSON event line with tenant_id injected for Splunk indexing.
func PrintJSONEvent(v interface{}, tenantID string) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("could not marshal event: %w", err)
	}

	var event map[string]interface{}
	if err := json.Unmarshal(raw, &event); err != nil {
		return fmt.Errorf("could not unmarshal event: %w", err)
	}
	event["tenant_id"] = tenantID

	out, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal event with tenant_id: %w", err)
	}
	fmt.Println(string(out))
	return nil
}
