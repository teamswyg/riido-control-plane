package authpep

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const agentReadFixtureService = "synthetic-distinct-agent-read-service-token"

func agentReadFixtureToken() string {
	return riidoaiserver.CLIAgentReadTokenPrefix + strings.Repeat("A", 43)
}

func agentReadFixtureRequest() AuthorizationRequest {
	r := threadReadFixtureRequest()
	r.Resource = riidoaiserver.AuthorizationResourceCLIAgentRead
	return r
}

func agentReadFixtureResponse() cliThreadReadResponse {
	r := threadReadFixtureResponse()
	r.Schema, r.Purpose = "riido-cli-agent-read.v1", "task_agents_read"
	return r
}

func newAgentReadFixture(t *testing.T, handler http.HandlerFunc) (*CLIThreadReadAuthorizer, *httptest.Server) {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	a, err := NewCLIAgentReadAuthorizer(CLIThreadReadConfig{MonolithOrigin: server.URL, Resource: threadReadFixtureResource, ServiceToken: agentReadFixtureService, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return a, server
}

func assertAgentReadFixtureRequest(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Method != http.MethodPost || r.URL.Path != "/internal/cli/agent-read/authorize" || r.Header.Get("Authorization") != "Bearer "+agentReadFixtureService || r.Header.Get("Cookie") != "" {
		t.Error("invalid agent service request")
	}
	var body map[string]string
	if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 5 || body["ticket"] != agentReadFixtureToken() || body["resource"] != threadReadFixtureResource || body["purpose"] != "task_agents_read" || body["workspace_id"] != threadReadFixtureWorkspace || body["task_id"] != "task-fixture" {
		t.Error("invalid agent task-bound check")
	}
}

func agentDirectoryHTTP(t *testing.T, server *httptest.Server, suffix string) []byte {
	t.Helper()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/v1/cli/workspaces/"+threadReadFixtureWorkspace+"/tasks/task-fixture/agents"+suffix, nil)
	req.Header.Set("Authorization", "Bearer "+agentReadFixtureToken())
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal("CP HTTP failed", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusOK || response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("agent directory HTTP failed", response.StatusCode, err)
	}
	return raw
}
