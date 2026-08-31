package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExternalHTTPAuthorizerUsesConfiguredTimeoutForDefaultClient(t *testing.T) {
	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{
		Endpoint: "https://authorizer.example.com", Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
	}
	if authorizer.timeout != 5*time.Second || authorizer.httpClient.Timeout != 5*time.Second {
		t.Fatalf("timeouts = (%s, %s), want 5s", authorizer.timeout, authorizer.httpClient.Timeout)
	}
}

func TestExternalHTTPAuthorizerTracesOutcomeWithoutRequestSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Traceparent"); got != "outgoing" {
			t.Fatalf("traceparent = %q", got)
		}
		_ = json.NewEncoder(w).Encode(externalAuthorizerResponse{
			SchemaVersion: ExternalAuthorizerResponseSchemaVersion, Allowed: true, PrincipalID: "user-1",
		})
	}))
	defer server.Close()
	authorizer, err := NewExternalHTTPAuthorizer(ExternalHTTPAuthorizerConfig{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("NewExternalHTTPAuthorizer: %v", err)
	}
	trace := &propagatingTraceRecorder{recordingTraceRecorder: &recordingTraceRecorder{}}
	_, err = authorizer.Authorize(WithTraceRecorder(context.Background(), trace), "secret-token", AuthorizationRequest{
		Resource: AuthorizationResourceAIAgentClient, Action: AuthorizationActionDeviceRead, WorkspaceID: "workspace-secret",
	})
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	spans := trace.snapshot()
	if len(spans) != 1 || spans[0].Name != "external_authorizer.authorize" || !spans[0].Ended {
		t.Fatalf("spans = %+v", spans)
	}
	if got := spans[0].Attributes["riido.external_authorizer.outcome"]; got != "allowed" {
		t.Fatalf("outcome = %q", got)
	}
	for _, value := range spans[0].Attributes {
		if value == "secret-token" || value == "workspace-secret" {
			t.Fatalf("trace attributes contain request secret: %+v", spans[0].Attributes)
		}
	}
}
