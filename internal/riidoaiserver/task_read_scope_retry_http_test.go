package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPTaskAssignmentRecoversScopeReadAndReplaysSameRequest(t *testing.T) {
	client, initial, _ := taskScopeRetryFixture(t, http.StatusServiceUnavailable, 1, 0)
	agents := NewDevelopmentAIAgentClientStore()
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: agents})
	t.Cleanup(func() { store.Close() })
	server := NewServer(ServerConfig{
		AIAgentClient: agents, Assignment: store, TaskContext: client,
		Authorizer: aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, "user-1"),
	}).Handler()
	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-private/agent-assignments"
	body := `{"agent_id":"agent-owned-codex","request_id":"request-scope-retry","instruction":"Explain this task"}`
	var first AIAgentTaskActionResponse
	for i := range 2 {
		resp := serveAssignTaskBoundary(server, path, "ai-agent-token", body)
		if resp.Code != http.StatusAccepted {
			t.Fatalf("attempt %d status=%d body=%s", i, resp.Code, resp.Body.String())
		}
		var got AIAgentTaskActionResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = got
		} else if got.AssignmentID != first.AssignmentID || got.RunID != first.RunID || got.ThreadID != first.ThreadID {
			t.Fatalf("replay created another execution: first=%+v got=%+v", first, got)
		}
	}
	if first.AssignmentID == "" || initial.Load() != 2 {
		t.Fatalf("missing assignment or repeated capture: assignment=%q calls=%d", first.AssignmentID, initial.Load())
	}
}
