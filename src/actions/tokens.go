package actions

import (
	"context"
	"fmt"

	"go.1password.io/eventsapi-splunk/splunk"
	"go.1password.io/eventsapi-splunk/utils"
)

func GetEventsToken(ctx context.Context, splunkAPI *splunk.SplunkAPI, secretName string) (string, error) {
	splunkRes, err := splunkAPI.GetPasswords(ctx, secretName, utils.SecretRealm)
	if err != nil {
		err := fmt.Errorf("GetEventsToken call to splunk failed: %w", err)
		return "", err
	}
	if len(splunkRes.Entry.Content.Dict.Key) == 0 {
		err := fmt.Errorf("splunk response is missing credentials")
		return "", err
	}

	return splunkRes.Entry.Content.Dict.Key[0].Text, nil
}

func CreateEventsToken(ctx context.Context, splunkAPI *splunk.SplunkAPI, secretName, authToken string) error {
	err := splunkAPI.CreatePassword(ctx, secretName, authToken, utils.SecretRealm)
	if err != nil {
		err := fmt.Errorf("CreateEventsToken call to splunk failed: %w", err)
		return err
	}
	return nil
}
