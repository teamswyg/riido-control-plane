package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newDirectCallFixture(t *testing.T, path string, legacy bool) *commentCallFixture {
	t.Helper()
	f := &commentCallFixture{snapshots: &memoryAIAgentClientSnapshotStore{}, status: 200}
	f.client = openPersistentAIAgentClientStoreForDelegateTest(t, f.snapshots)
	f.evidence = verifiedCommentDelivery{TaskID: "handoff-task", WorkspaceID: defaultAIAgentClientWorkspaceID, PrincipalID: "user-1", Context: AIAgentTaskContext{TeamID: "team-direct", Component: AIAgentTaskContextComponent{ID: "handoff-task", ComponentType: "task", Title: "Implement a greeting"}, Document: AIAgentTaskContextDocument{Content: "Implement a Go greeting and request a code review."}}}
	key := strings.Repeat("k", 32)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get(CommentDeliveryKeyHeader) != key || r.URL.Path != "/internal/agent-assignment-context/"+defaultAIAgentClientWorkspaceID+"/team-direct/handoff-task/user-1" {
			t.Error("unexpected evidence lookup", r.Method, r.URL.Path)
			w.WriteHeader(403)
			return
		}
		if f.status != 200 {
			w.WriteHeader(f.status)
			return
		}
		if err := json.NewEncoder(w).Encode(f.evidence); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(endpoint.Close)
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "task:handoff-task:assign"}}, {PrincipalID: "event-agent", Token: "assignment-token", Scopes: []string{"agent:agent-public-openclaw:events:write", "agent:agent-owned-codex:events:write"}}})
	if err != nil {
		t.Fatal(err)
	}
	f.assignments = &commentCompletionStore{Store: NewStoreWithConfig(StoreConfig{AgentRegistry: f.client})}
	t.Cleanup(f.assignments.Close)
	f.s = NewServer(ServerConfig{AIAgentClient: f.client, Assignment: f.assignments, AssignmentContext: NewMemoryAssignmentContextStore(), TaskContext: &assignmentHTTPRequestTaskContextReader{contextSnapshot: f.evidence.Context}, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: endpoint.URL, Key: key}, Authorizer: auth})
	f.s.commentReactions = nil
	if legacy {
		f.s.config.CommentDelivery.Key = ""
	}
	r := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/"+defaultAIAgentClientWorkspaceID+"/ai-agent/tasks/handoff-task/"+path, strings.NewReader(`{"agent_id":"agent-public-openclaw","request_id":"direct-greet","instruction":"Implement the Go greeting and have its output independently reviewed."}`))
	r.Header.Set(aiAgentTokenHeader, "user-token")
	w := httptest.NewRecorder()
	f.s.Handler().ServeHTTP(w, r)
	if w.Code != 202 {
		t.Fatalf("assignment: %d %s", w.Code, w.Body.String())
	}
	if err = json.Unmarshal(w.Body.Bytes(), &f.source); err != nil {
		t.Fatal(err)
	}
	f.s.config.CommentDelivery.Key = key
	return f
}
