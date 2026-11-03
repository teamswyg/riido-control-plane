package authpep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const (
	threadReadFixtureService   = "synthetic-service-token-for-thread-reads"
	threadReadFixtureResource  = "https://cp.example.test"
	threadReadFixtureWorkspace = "workspace-dev-riid"
)

func threadReadFixtureToken() string {
	return riidoaiserver.CLIThreadReadTokenPrefix + strings.Repeat("A", 43)
}

func threadReadFixtureRequest() AuthorizationRequest {
	return AuthorizationRequest{Resource: riidoaiserver.AuthorizationResourceCLIThreadRead, Action: AuthorizationActionRead, WorkspaceID: threadReadFixtureWorkspace, TaskID: "task-fixture"}
}

func threadReadFixtureResponse() cliThreadReadResponse {
	now := time.Now().UTC()
	return cliThreadReadResponse{Schema: cliThreadReadSchema, Allowed: true, PrincipalID: "user-1", Roles: []string{}, Resource: threadReadFixtureResource, Purpose: cliThreadReadPurpose, WorkspaceID: threadReadFixtureWorkspace, TaskID: "task-fixture", ExpiresAt: now.Add(55 * time.Second), CheckedAt: now}
}

func newThreadReadFixture(t *testing.T, handler http.HandlerFunc) (*CLIThreadReadAuthorizer, *httptest.Server) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	a, err := NewCLIThreadReadAuthorizer(CLIThreadReadConfig{MonolithOrigin: server.URL, Resource: threadReadFixtureResource, ServiceToken: threadReadFixtureService, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return a, server
}

func assertThreadReadFixtureRequest(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/internal/cli/thread-read/authorize" || r.Header.Get("Authorization") != "Bearer "+threadReadFixtureService || r.Header.Get("Cookie") != "" {
		t.Error("invalid service request")
	}
	var body map[string]string
	if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 5 || body["ticket"] != threadReadFixtureToken() || body["resource"] != threadReadFixtureResource || body["purpose"] != cliThreadReadPurpose || body["workspace_id"] != threadReadFixtureWorkspace || body["task_id"] != "task-fixture" {
		t.Error("invalid task-bound check")
	}
}
