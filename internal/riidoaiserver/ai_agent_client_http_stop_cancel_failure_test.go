package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientStopDoesNotMutateThreadWhenDurableCancelFails(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}

	assigned, err := aiAgentStore.AssignAIAgentTask(ctx, principal, "task-stop-cancel-fails", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-missing-durable",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if assigned.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("test setup assignment should be running: %+v", assigned)
	}

	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, principal.PrincipalID),
	}).Handler()
	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stop-cancel-fails/stop", strings.NewReader(`{"agent_id":"agent-public-openclaw","reason":"user requested stop"}`))
	stopReq.Header.Set("Authorization", "Bearer ai-agent-token")
	stopResp := httptest.NewRecorder()
	server.ServeHTTP(stopResp, stopReq)
	if stopResp.Code == http.StatusAccepted {
		t.Fatalf("stop should fail when durable cancel fails: status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}
	if !strings.Contains(stopResp.Body.String(), "assignment asn-missing-durable not found") {
		t.Fatalf("stop error should expose durable cancel failure, status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}

	threads, err := aiAgentStore.ListAIAgentTaskThreads(ctx, principal, "task-stop-cancel-fails")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads)
	}
	thread := threads.Threads[0]
	if !taskThreadHasActiveStream(thread) ||
		thread.AssignmentState != AgentAssignmentStateRunning ||
		thread.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("failed durable cancel must not pre-stop read model thread: %+v", thread)
	}
}
