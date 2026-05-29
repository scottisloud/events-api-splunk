package actions

import (
	"context"
	"time"

	"go.1password.io/eventsapi-splunk/api"
)

func StartItemUsages(cursorFile string, limit int, startAt *time.Time, eventsAPI *api.EventsAPI) {
	pollEvents("ItemUsages", cursorFile, limit, startAt, eventsAPI,
		func(ctx context.Context, body interface{}) (eventBatch, error) {
			return eventsAPI.ItemUsagesRequest(ctx, body)
		})
}
