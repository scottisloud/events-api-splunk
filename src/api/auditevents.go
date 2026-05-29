package api

import (
	"context"
	"fmt"
	"time"
)

type AuditEvent struct {
	UUID          string                     `json:"uuid"`
	Timestamp     time.Time                  `json:"timestamp"`
	ActorUUID     string                     `json:"actor_uuid"`
	ActorDetails  *AuditEventResourceDetails `json:"actor_details,omitempty"`
	Action        string                     `json:"action"`
	ObjectType    string                     `json:"object_type"`
	ObjectUUID    string                     `json:"object_uuid"`
	ObjectDetails *AuditEventResourceDetails `json:"object_details,omitempty"`
	AuxID         int64                      `json:"aux_id,omitempty"`
	AuxUUID       string                     `json:"aux_uuid,omitempty"`
	AuxDetails    *AuditEventResourceDetails `json:"aux_details,omitempty"`
	AuxInfo       string                     `json:"aux_info,omitempty"`

	Session  AuditEventSession   `json:"session"`
	Location *AuditEventLocation `json:"location,omitempty"`
}

type AuditEventResourceDetails struct {
	UUID  string `json:"uuid,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type AuditEventSession struct {
	UUID       string    `json:"uuid"`
	LoginTime  time.Time `json:"login_time"`
	DeviceUUID string    `json:"device_uuid"`
	IP         string    `json:"ip"`
}

type AuditEventLocation struct {
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type AuditEventsResponse struct {
	*CursorResponse
	AuditEvents []AuditEvent `json:"items"`
}

func (s *AuditEventsResponse) EventCount() int { return len(s.AuditEvents) }

func (s *AuditEventsResponse) PrintEvents(tenantID string) error {
	for i, v := range s.AuditEvents {
		if err := PrintJSONEvent(v, tenantID); err != nil {
			return fmt.Errorf("could not print event: %d, error: %w", i, err)
		}
	}
	return nil
}

func (e *EventsAPI) AuditEventsRequest(ctx context.Context, body interface{}) (*AuditEventsResponse, error) {
	return postEvents[AuditEventsResponse](e, ctx, "/api/v1/auditevents", body)
}
