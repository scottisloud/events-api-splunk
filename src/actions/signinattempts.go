package actions

import (
	"context"
	"time"

	"go.1password.io/eventsapi-splunk/api"
)

func StartSignIns(cursorFile string, limit int, startAt *time.Time, eventsAPI *api.EventsAPI) {
	pollEvents("SignIns", cursorFile, limit, startAt, eventsAPI,
		func(ctx context.Context, body interface{}) (eventBatch, error) {
			return eventsAPI.SignInAttemptsRequest(ctx, body)
		})
}
