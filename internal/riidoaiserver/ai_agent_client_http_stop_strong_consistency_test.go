package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
)

func TestHTTPAIAgentClientStopAgentStronglyClosesReadModelBeforeDaemonAck(t *testing.T) {
	ctx := context.Background()
	const token = "owner-token"
	const taskID = "task-stop-strong-agent"
	const agentID = "agent-public-openclaw"
	base := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent"
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(assignmentStore.Close)
	server := newStopStrongConsistencyServer(t, aiAgentStore, assignmentStore, token)

	first := createProviderMultiAssignment(t, server, base, token, taskID, agentID)
	assertProviderMultiPollStart(t, server, token, first, "daemon-shared-studio", "device-shared-studio", "runtime-openclaw-shared")

	stopBytes := aiAgentSmokeRequest(t, server, http.MethodPost, base+"/tasks/"+taskID+"/agent-assignments/"+agentID+"/stop", token, `{"reason":"user stop"}`, http.StatusAccepted)
	var stopped AIAgentTaskActionResponse
	aiAgentSmokeDecode(t, stopBytes, &stopped)
	if stopped.AssignmentID != first.AssignmentID ||
		stopped.AssignmentState != AgentAssignmentStateStopped ||
		stopped.WorkStatus != AgentWorkStatusIdle ||
		stopped.ActiveStream != nil {
		t.Fatalf("stop response = %+v first=%+v", stopped, first)
	}

	assertAssignmentProjectionState(t, ctx, assignmentStore, first.AssignmentID, AssignmentCancelling)
	assertStoppedThreads(t, server, base, token, taskID, first.AssignmentID)
}

func newStopStrongConsistencyServer(t *testing.T, aiAgentStore *DevelopmentAIAgentClientStore, assignmentStore *Store, token string) http.Handler {
	t.Helper()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "agent:*:poll"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	return NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()
}

func assertAssignmentProjectionState(t *testing.T, ctx context.Context, store *Store, assignmentID string, want AssignmentState) {
	t.Helper()
	projection, ok, err := store.LoadAssignmentProjection(ctx, assignmentID)
	if err != nil || !ok || projection.Assignment.State != want {
		t.Fatalf("projection %s = %+v ok=%v err=%v want=%s", assignmentID, projection, ok, err, want)
	}
}
