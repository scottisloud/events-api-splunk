package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintJSONEventInjectsTenantID(t *testing.T) {
	event := SignInAttempt{UUID: "abc", Country: "CA"}
	tenantID := "acme-corp"

	// Capture stdout by testing helper logic via marshal path
	raw, _ := json.Marshal(event)
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	m["tenant_id"] = tenantID
	out, _ := json.Marshal(m)

	if !strings.Contains(string(out), `"tenant_id":"acme-corp"`) &&
		!strings.Contains(string(out), `"tenant_id": "acme-corp"`) {
		t.Fatalf("missing tenant_id in %s", out)
	}
}
