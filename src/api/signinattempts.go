package api

import (
	"context"
	"fmt"
)

type SignInAttempt struct {
	UUID        string                  `json:"uuid"`
	SessionUUID string                  `json:"session_uuid"`
	Timestamp   FixedFormatTime         `json:"timestamp"`
	Country     string                  `json:"country"`
	Category    string                  `json:"category"`
	Type        string                  `json:"type"`
	Details     *SignInAttemptDetails   `json:"details"`
	TargetUser  SignInAttemptTargetUser `json:"target_user"`
	Client      SignInAttemptClient     `json:"client"`
	Location    *SignInAttemptLocation  `json:"location,omitempty"`
}

type SignInAttemptDetails struct {
	Value string `json:"value"`
}
type SignInAttemptTargetUser struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type SignInAttemptClient struct {
	AppName         string `json:"app_name"`
	AppVersion      string `json:"app_version"`
	PlatformName    string `json:"platform_name"`
	PlatformVersion string `json:"platform_version"`
	OSName          string `json:"os_name"`
	OSVersion       string `json:"os_version"`
	IPAddress       string `json:"ip_address"`
}

type SignInAttemptLocation struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CursorResponse struct {
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"has_more"`
}

// GetCursor and GetHasMore expose the embedded cursor fields through the
// eventBatch interface used by the shared poller in the actions package.
func (c *CursorResponse) GetCursor() string { return c.Cursor }
func (c *CursorResponse) GetHasMore() bool  { return c.HasMore }

type SignInAttemptResponse struct {
	*CursorResponse
	Items []SignInAttempt `json:"items"`
}

func (s *SignInAttemptResponse) EventCount() int { return len(s.Items) }

func (s *SignInAttemptResponse) PrintEvents(tenantID string) error {
	for i, v := range s.Items {
		if err := PrintJSONEvent(v, tenantID); err != nil {
			return fmt.Errorf("could not print event: %d, error: %w", i, err)
		}
	}
	return nil
}

func (e *EventsAPI) SignInAttemptsRequest(ctx context.Context, body interface{}) (*SignInAttemptResponse, error) {
	return postEvents[SignInAttemptResponse](e, ctx, "/api/v1/signinattempts", body)
}
