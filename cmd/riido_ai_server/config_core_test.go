package main

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvParsesAddressShutdownAndStaticAuthorizer(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAddr, ":9090")
	t.Setenv(envShutdownTimeoutSeconds, "7")
	t.Setenv(envAuthzTokensJSON, `[{
		"principal_id":"daemon:agent-a",
		"token":"static-token",
		"scopes":["agent:agent-a:poll"]
	}]`)

	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.Addr != ":9090" || config.ShutdownTimeout != 7*time.Second {
		t.Fatalf("config = %+v", config)
	}
	_, err = config.Authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "agent-a",
	})
	if err != nil {
		t.Fatalf("static authorize: %v", err)
	}
}

func TestConfigFromEnvDefaultsToPublicHealthOnlyRuntime(t *testing.T) {
	clearRiidoAIServerEnv(t)
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	if config.Addr != ":8080" || config.ShutdownTimeout != 10*time.Second {
		t.Fatalf("defaults = %+v", config)
	}
	if config.MetricsLogInterval != 0 || config.Authorizer != nil {
		t.Fatalf("optional config should be disabled: %+v", config)
	}
}

func TestConfigFromEnvParsesWebAllowedOrigins(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envWebAllowedOrigins, " https://app.riido.io, http://localhost:5173/ , https://app.riido.io ")
	config, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv: %v", err)
	}
	want := []string{"https://app.riido.io", "http://localhost:5173"}
	if !reflect.DeepEqual(config.WebAllowedOrigins, want) {
		t.Fatalf("web origins = %v, want %v", config.WebAllowedOrigins, want)
	}
}

func TestParseWebAllowedOriginsRejectsInvalidOrigins(t *testing.T) {
	for _, value := range []string{"*", "ftp://app.riido.io", "https://app.riido.io/path", "https://app.riido.io?debug=true", "https://user@app.riido.io"} {
		t.Run(value, func(t *testing.T) {
			if _, err := parseWebAllowedOrigins(value); err == nil || !strings.Contains(err.Error(), envWebAllowedOrigins) {
				t.Fatalf("parseWebAllowedOrigins err=%v", err)
			}
		})
	}
}
