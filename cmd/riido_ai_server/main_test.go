package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func TestConfigFromEnvParsesAddressShutdownRegistryAndStaticAuthorizer(t *testing.T) {
	clearRiidoAIServerEnv(t)
	t.Setenv(envAddr, ":9090")
	t.Setenv(envShutdownTimeoutSeconds, "7")
	t.Setenv(envAgentBindingsJSON, `[{
		"agent_id":"agent-a",
		"daemon_id":"daemon-a",
		"device_id":"device-a",
		"runtime_id":"runtime-a",
		"runtime_provider":"codex"
	}]`)
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
	binding, ok := config.AgentRegistry.LookupAgent("agent-a")
	if !ok || binding.RuntimeID != "runtime-a" || binding.RuntimeProvider != "codex" {
		t.Fatalf("binding = %+v ok=%v", binding, ok)
	}
	if _, err := config.Authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "agent-a",
	}); err != nil {
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
	if config.AgentRegistry != nil || config.Authorizer != nil {
		t.Fatalf("optional config should be nil: %+v", config)
	}
}

func TestParseAgentRegistryJSONRejectsUnknownField(t *testing.T) {
	if _, err := parseAgentRegistryJSON(`[{"agent_id":"agent-a","daemon_id":"daemon-a","runtime_id":"runtime-a","runtime_provider":"codex","extra":true}]`); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestParseAuthzTokensJSONRejectsUnknownField(t *testing.T) {
	if _, err := parseAuthzTokensJSON(`[{"principal_id":"user-a","token":"static-token","scopes":["riido:*"],"extra":true}]`); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestAuthorizerFromEnvFallsBackFromStaticToExternalOnlyWhenUnauthenticated(t *testing.T) {
	clearRiidoAIServerEnv(t)
	var externalCalls atomic.Int32
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		externalCalls.Add(1)
		var req struct {
			SchemaVersion string `json:"schema_version"`
			BearerToken   string `json:"bearer_token"`
			Request       struct {
				Resource string `json:"resource"`
				Action   string `json:"action"`
			} `json:"request"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.SchemaVersion != riidoaiserver.ExternalAuthorizerRequestSchemaVersion || req.BearerToken != "external-token" || req.Request.Resource != "metrics" || req.Request.Action != "read" {
			t.Fatalf("external request = %+v", req)
		}
		_ = json.NewEncoder(w).Encode(struct {
			SchemaVersion string `json:"schema_version"`
			Allowed       bool   `json:"allowed"`
			PrincipalID   string `json:"principal_id"`
		}{
			SchemaVersion: riidoaiserver.ExternalAuthorizerResponseSchemaVersion,
			Allowed:       true,
			PrincipalID:   "external-user",
		})
	}))
	defer authorizerServer.Close()

	t.Setenv(envAuthzTokensJSON, `[{"principal_id":"static-user","token":"static-token","scopes":["metrics:read"]}]`)
	t.Setenv(envExternalAuthzURL, authorizerServer.URL)
	t.Setenv(envExternalAuthzTimeout, "1")
	authorizer, err := authorizerFromEnv()
	if err != nil {
		t.Fatalf("authorizerFromEnv: %v", err)
	}

	if _, err := authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceMetrics,
		Action:   riidoaiserver.AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("static authorize: %v", err)
	}
	if got := externalCalls.Load(); got != 0 {
		t.Fatalf("external calls after static token = %d", got)
	}
	if _, err := authorizer.Authorize(context.Background(), "external-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceMetrics,
		Action:   riidoaiserver.AuthorizationActionRead,
	}); err != nil {
		t.Fatalf("external authorize: %v", err)
	}
	if got := externalCalls.Load(); got != 1 {
		t.Fatalf("external calls after fallback = %d", got)
	}
	if _, err := authorizer.Authorize(context.Background(), "static-token", riidoaiserver.AuthorizationRequest{
		Resource: riidoaiserver.AuthorizationResourceAgent,
		Action:   riidoaiserver.AuthorizationActionPoll,
		AgentID:  "agent-a",
	}); err == nil {
		t.Fatal("expected forbidden static scope to stop fallback")
	}
	if got := externalCalls.Load(); got != 1 {
		t.Fatalf("external should not run after forbidden static scope, calls=%d", got)
	}
}

func TestEnvDurationSecondsRejectsNonPositiveValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv(envShutdownTimeoutSeconds, value)
			if _, err := envDurationSeconds(envShutdownTimeoutSeconds, time.Second); err == nil || !strings.Contains(err.Error(), envShutdownTimeoutSeconds) {
				t.Fatalf("envDurationSeconds err=%v", err)
			}
		})
	}
}

func clearRiidoAIServerEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		envAddr,
		envShutdownTimeoutSeconds,
		envAgentBindingsJSON,
		envAuthzTokensJSON,
		envExternalAuthzURL,
		envExternalAuthzAudience,
		envExternalAuthzTimeout,
	} {
		t.Setenv(key, "")
	}
}
