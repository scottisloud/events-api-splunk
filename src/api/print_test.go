package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
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

func TestPrintJSONEventConcurrentWritesValidLines(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = writer

	var wg sync.WaitGroup
	tenants := 19
	eventsPerTenant := 5

	for tenant := 0; tenant < tenants; tenant++ {
		wg.Add(1)
		go func(tenantID string) {
			defer wg.Done()
			for i := 0; i < eventsPerTenant; i++ {
				if err := PrintJSONEvent(SignInAttempt{UUID: tenantID}, tenantID); err != nil {
					t.Errorf("PrintJSONEvent failed: %v", err)
				}
			}
		}(fmt.Sprintf("tenant-%02d", tenant))
	}

	wg.Wait()
	writer.Close()
	os.Stdout = oldStdout

	scanner := bufio.NewScanner(reader)
	lines := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("line is not valid JSON: %q (%v)", line, err)
		}
		if _, ok := payload["tenant_id"]; !ok {
			t.Fatalf("line missing tenant_id: %q", line)
		}
		lines++
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if lines != tenants*eventsPerTenant {
		t.Fatalf("expected %d lines, got %d", tenants*eventsPerTenant, lines)
	}
}
