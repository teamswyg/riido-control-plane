package riidoaiserver

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const assignmentContextTestSecret = "device-secret-dev"

type assignmentContextFixture struct {
	t        *testing.T
	server   http.Handler
	contexts *MemoryAssignmentContextStore
	leased   Assignment
}

// newAssignmentContextFixture assigns task-private through the client API with
// a request-scoped task context, then leases it with the daemon poll.
func newAssignmentContextFixture(t *testing.T, contexts *MemoryAssignmentContextStore) *assignmentContextFixture {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "task:task-private:assign", "agent:agent-owned-codex:poll"}}})
	if err != nil {
		t.Fatal(err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	aiAgentStore.deviceCredentials["device-dev-macbook"] = deviceCredentialRecord{deviceID: "device-dev-macbook", secretHash: sha256.Sum256([]byte(assignmentContextTestSecret)), ownerPrincipalID: "user-1", workspaceID: "workspace-dev-riid"}
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(func() { assignmentStore.Close() })
	config := ServerConfig{AIAgentClient: aiAgentStore, Assignment: assignmentStore, TaskContext: &assignmentHTTPRequestTaskContextReader{contextSnapshot: assignmentContextTestTaskContext()}, Authorizer: authorizer}
	if contexts != nil {
		config.AssignmentContext = contexts
	}
	f := &assignmentContextFixture{t: t, server: NewServer(config).Handler(), contexts: contexts}
	if code, body := f.do("/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-private/assignment", `{"agent_id":"agent-owned-codex","request_id":"req-assign-1"}`, func(r *http.Request) { r.Header.Set(aiAgentTokenHeader, "user-token") }); code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", code, body)
	}
	code, body := f.do("/v1/agents/agent-owned-codex/poll", `{"daemon_id":"daemon-dev-macbook","device_id":"device-dev-macbook","runtime_id":"runtime-codex-dev"}`, f.device)
	var poll PollResponse
	if code != http.StatusOK || json.Unmarshal([]byte(body), &poll) != nil || poll.Assignment == nil || poll.Assignment.LeaseToken == "" {
		t.Fatalf("poll status=%d body=%s", code, body)
	}
	f.leased = *poll.Assignment
	return f
}

func (f *assignmentContextFixture) device(r *http.Request) {
	r.Header.Set(deviceIDHeader, "device-dev-macbook")
	r.Header.Set(deviceSecretHeader, assignmentContextTestSecret)
}

func (f *assignmentContextFixture) do(path, body string, decorate func(*http.Request)) (int, string) {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if decorate != nil {
		decorate(req)
	}
	resp := httptest.NewRecorder()
	f.server.ServeHTTP(resp, req)
	return resp.Code, resp.Body.String()
}

func (f *assignmentContextFixture) read(view, target, lease string) (int, string) {
	body, _ := json.Marshal(AgentAssignmentContextRequest{AssignmentID: f.leased.ID, LeaseToken: lease, View: view, Target: target, RequestID: "cli_1"})
	return f.do("/v1/agents/agent-owned-codex/assignment-context", string(body), f.device)
}
