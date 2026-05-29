package actions

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.1password.io/eventsapi-splunk/api"
	"go.1password.io/eventsapi-splunk/store"
)

// eventBatch is the response shape shared by every Events API endpoint poller.
// The three feeds (sign-in attempts, item usages, audit events) differ only in
// the request they issue and the slice they decode into, so they all satisfy
// this interface and run through pollEvents below.
type eventBatch interface {
	PrintEvents(tenantID string) error
	GetCursor() string
	GetHasMore() bool
	EventCount() int
}

// requestFunc fetches a single page of events for the given request body.
type requestFunc func(ctx context.Context, body interface{}) (eventBatch, error)

// pollEvents drives one event feed: it resets the cursor on first run, then
// continuously polls for new events, printing them and persisting the cursor
// after each batch, until interrupted.
func pollEvents(feedName, cursorFile string, limit int, startAt *time.Time, eventsAPI *api.EventsAPI, request requestFunc) {
	log.Printf("Starting %s...", feedName)

	cursorStore, err := store.OpenStore(cursorFile)
	if err != nil {
		panic(fmt.Errorf("could not open cursor file: %w", err))
	}
	defer cursorStore.CloseStore()

	cursor, err := cursorStore.GetCursor()
	if err != nil {
		panic(fmt.Errorf("could not read cursor: %w", err))
	}

	// Set up notify channel so that we can gracefully shut down.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cursor == "" {
		log.Println("Performing cursor reset")
		res, err := request(ctx, api.CursorResetRequest{Limit: limit, StartTime: startAt})
		if err != nil {
			panic(fmt.Errorf("%s cursor reset request failed: %w", feedName, err))
		}
		if err := res.PrintEvents(eventsAPI.TenantID); err != nil {
			panic(fmt.Errorf("PrintEvents failed: %w", err))
		}
		if err := cursorStore.SaveCursor(res.GetCursor()); err != nil {
			panic(err)
		}
		cursor = res.GetCursor()
	} else {
		log.Println("Using stored cursor")
	}

	for {
		select {
		case <-sigCh:
			log.Println("Interrupted, shutting down")
			cancel()
			if err := cursorStore.CloseStore(); err != nil {
				log.Println(fmt.Errorf("could not close store: %w", err))
				os.Exit(1)
			}
			log.Println("Gracefully shutdown")
			os.Exit(0)
		default:
			res, err := request(ctx, api.CursorRequest{Cursor: cursor})
			if err != nil {
				log.Printf("%s request failed: %s\n", feedName, err)
				time.Sleep(30 * time.Second)
				continue
			}

			if res.EventCount() == 0 && !res.GetHasMore() {
				// Don't bother printing or storing this cursor,
				// we will reuse the last one until we receive some events.
				time.Sleep(10 * time.Second)
				continue
			}

			if err := res.PrintEvents(eventsAPI.TenantID); err != nil {
				panic(fmt.Errorf("PrintEvents failed: %w", err))
			}
			if err := cursorStore.SaveCursor(res.GetCursor()); err != nil {
				panic(fmt.Errorf("SaveCursor failed: %w", err))
			}
			cursor = res.GetCursor()
		}
	}
}
