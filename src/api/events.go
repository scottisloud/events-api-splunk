package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type EventsAPI struct {
	client    *http.Client
	AuthToken string
	BaseUrl   string
	TenantID  string
}

var Version string
var DefaultUserAgent = fmt.Sprintf("1Password Events API for Splunk / %s", Version)

func NewEventsAPI(authToken string, url string) *EventsAPI {
	log.Println("New Events API Version:", Version)
	retryHTTPClient := retryablehttp.NewClient()
	retryHTTPClient.Logger = &loggerWrapper{}

	client := &EventsAPI{
		client:    retryHTTPClient.StandardClient(),
		AuthToken: authToken,
		BaseUrl:   url,
	}
	return client
}

// postEvents performs a POST to an Events API endpoint and decodes the JSON
// response into *T. The three event endpoints (sign-in attempts, item usages,
// audit events) share identical request/decode handling and differ only in the
// route and response type, so they all funnel through here.
func postEvents[T any](e *EventsAPI, ctx context.Context, route string, body interface{}) (*T, error) {
	res, err := e.request(ctx, "POST", route, body)
	if err != nil {
		return nil, fmt.Errorf("could not make EventAPIRequest: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	out := new(T)
	if err := json.Unmarshal(resBody, out); err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %s", string(resBody))
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received a non 200 response: %v", string(resBody))
	}

	return out, nil
}

func (e *EventsAPI) request(ctx context.Context, method string, route string, body interface{}) (*http.Response, error) {
	var b io.Reader
	if body != nil {
		reqBody, err := json.Marshal(body)
		if err != nil {
			err := fmt.Errorf("could not marshal request: %w", err)
			panic(err)
		}
		b = bytes.NewReader(reqBody)
	}
	req, err := http.NewRequestWithContext(ctx, method, e.BaseUrl+route, b)
	if err != nil {
		err := fmt.Errorf("could not create new request: %w", err)
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", e.AuthToken))
	req.Header.Add("User-Agent", DefaultUserAgent)
	res, err := e.client.Do(req)
	if err != nil {
		err := fmt.Errorf("could not make request: %w", err)
		return nil, err
	}
	return res, nil
}

type CursorRequest struct {
	Cursor string `json:"cursor"`
}
type CursorResetRequest struct {
	Limit     int        `json:"limit"`
	StartTime *time.Time `json:"start_time,omitempty"`
}

type loggerWrapper struct {
}

func (l *loggerWrapper) Printf(s string, i ...interface{}) {
	log.Printf(s, i...)
}
