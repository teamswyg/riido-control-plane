package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// A failed push after an earlier success must not block repair by later history reads.
func TestAgentCommentCountRepairRetriesAfterFailedPushFollowingSuccess(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "repair-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", Instruction: "summarize this task"})
	if err != nil {
		t.Fatal(err)
	}
	var calls, isFailing atomic.Int32
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		if isFailing.Load() == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer endpoint.Close()
	server := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: endpoint.URL, Key: "key"}, AgentCommentCount: true})
	key := principal.WorkspaceID + "\x00" + action.TaskID
	server.pushAgentCommentCount(t.Context(), key, principal.WorkspaceID, action.TaskID)

	// The conversation changes but the monolith is temporarily unavailable.
	server.agentCommentCounts.markSent(key, -1)
	isFailing.Store(1)
	server.pushAgentCommentCount(t.Context(), key, principal.WorkspaceID, action.TaskID)
	isFailing.Store(0)

	before := calls.Load()
	server.repairAgentCommentCount(t.Context(), principal.WorkspaceID, action.TaskID)
	deadline := time.Now().Add(2 * time.Second)
	for calls.Load() == before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if calls.Load() == before {
		t.Fatal("history read did not retry the failed push")
	}
}
