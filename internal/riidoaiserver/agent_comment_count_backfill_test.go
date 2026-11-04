package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestAgentCommentCountBackfillPushesEveryTaskWithThreads(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "backfill-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", Instruction: "summarize this task"})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	tasks := map[string]string{}
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		tasks[body["task_id"].(string)] = body["workspace_id"].(string)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer endpoint.Close()
	server := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: endpoint.URL, Key: "key"}, AgentCommentCount: true})
	server.backfillAgentCommentCounts(t.Context())
	mu.Lock()
	defer mu.Unlock()
	if tasks[action.TaskID] != principal.WorkspaceID {
		t.Fatalf("backfill pushed %v, want %s in %s", tasks, action.TaskID, principal.WorkspaceID)
	}
}
