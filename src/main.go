package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"go.1password.io/eventsapi-splunk/actions"
	events "go.1password.io/eventsapi-splunk/api"
	"go.1password.io/eventsapi-splunk/config"
	"go.1password.io/eventsapi-splunk/splunk"
	"go.1password.io/eventsapi-splunk/utils"
)

var EventBuildType string

func main() {
	log.Println("Booting...")
	if EventBuildType == "" {
		panic(fmt.Errorf("missing EventBuildType flag"))
	}

	splunkHome := os.Getenv("SPLUNK_HOME")
	if splunkHome == "" {
		panic(fmt.Errorf("SPLUNK_HOME environment variable must be set"))
	}

	splunkEnv, err := config.NewSplunkEnv(splunkHome)
	if err != nil {
		panic(fmt.Errorf("could not create new splunk env: %w", err))
	}

	if err := splunkEnv.ConsumePendingCursorCleanups(); err != nil {
		log.Printf("cursor cleanup failed: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	splunkSession, _, err := reader.ReadLine()
	if err != nil {
		panic(fmt.Errorf("could not read session: %w", err))
	}

	splunkAPI := splunk.NewSplunkAPI(string(splunkSession))

	tenants, err := splunkEnv.LoadTenants()
	if err != nil {
		panic(fmt.Errorf("could not load tenants: %w", err))
	}

	for _, tenant := range tenants {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("tenant %q (key=%s) panic: %v", tenant.TenantID, tenant.TenantKey, r)
				}
			}()
			if err := processTenant(splunkEnv, splunkAPI, tenant); err != nil {
				log.Printf("tenant %q (key=%s) failed: %v", tenant.TenantID, tenant.TenantKey, err)
			}
		}()
	}
}

func processTenant(splunkEnv *config.SplunkEnv, splunkAPI *splunk.SplunkAPI, tenant config.TenantRuntime) error {
	log.Printf("Processing tenant %q (key=%s)", tenant.TenantID, tenant.TenantKey)

	cfg := tenant.Config
	var eventsToken string
	var err error

	if cfg.AuthToken != "" && tenant.TenantKey == utils.DefaultTenantKey {
		eventsToken = cfg.AuthToken
		err = actions.CreateEventsToken(context.TODO(), splunkAPI, tenant.SecretName, eventsToken)
		if err != nil {
			return fmt.Errorf("could not migrate token to storage passwords: %w", err)
		}
		splunkEnv.Config.AuthToken = ""
		if err := splunkEnv.UpdateConfig(splunkEnv.Config); err != nil {
			return fmt.Errorf("could not remove auth token from disk: %w", err)
		}
	} else {
		eventsToken, err = actions.GetEventsToken(context.TODO(), splunkAPI, tenant.SecretName)
		if err != nil {
			return fmt.Errorf("could not get token: %w", err)
		}
	}

	jwt, err := utils.ParseJWTClaims(eventsToken)
	if err != nil {
		return fmt.Errorf("could not parse jwt: %w", err)
	}
	if err := utils.ValidateTokenTenantKey(jwt, tenant.TenantKey); err != nil {
		return err
	}

	url := cfg.Url
	eventsURL, err := jwt.GetEventsURL()
	if err == nil {
		url = eventsURL
	}
	if url == "" {
		return fmt.Errorf("missing events API url for tenant %q", tenant.TenantID)
	}

	eventsAPI := events.NewEventsAPI(eventsToken, url)
	if _, err := eventsAPI.Introspect(context.TODO()); err != nil {
		return fmt.Errorf("token introspect failed for tenant %q: %w", tenant.TenantID, err)
	}
	eventsAPI.TenantID = tenant.TenantID
	startAt := cfg.StartAt

	if jwt.Features.Contains(utils.SignInAttemptsFeatureScope) && EventBuildType == utils.SignInAttemptsFeatureScope {
		cursorFile, err := config.ResolveCursorFile(splunkEnv.Home, cfg.SignInCursorFile)
		if err != nil {
			return fmt.Errorf("invalid sign-in cursor path: %w", err)
		}
		actions.StartSignIns(cursorFile, cfg.Limit, &startAt, eventsAPI)
	} else if jwt.Features.Contains(utils.ItemUsageFeatureScope) && EventBuildType == utils.ItemUsageFeatureScope {
		cursorFile, err := config.ResolveCursorFile(splunkEnv.Home, cfg.ItemUsageCursorFile)
		if err != nil {
			return fmt.Errorf("invalid item usage cursor path: %w", err)
		}
		actions.StartItemUsages(cursorFile, cfg.Limit, &startAt, eventsAPI)
	} else if jwt.Features.Contains(utils.AuditEventsFeatureScope) && EventBuildType == utils.AuditEventsFeatureScope {
		cursorFile, err := config.ResolveCursorFile(splunkEnv.Home, cfg.AuditEventsCursorFile)
		if err != nil {
			return fmt.Errorf("invalid audit events cursor path: %w", err)
		}
		actions.StartAuditEvents(cursorFile, cfg.Limit, &startAt, eventsAPI)
	} else {
		log.Printf(
			"tenant %q (key=%s): token missing scope %q for this input; skipping",
			tenant.TenantID, tenant.TenantKey, EventBuildType,
		)
	}

	return nil
}
