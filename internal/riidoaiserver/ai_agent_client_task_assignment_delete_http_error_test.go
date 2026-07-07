package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDeleteTaskAgentAssignmentRejectsUnscopedRouteAndBadJSON(t *testing.T) {
	const token = "user-token"
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       token,
		Scopes:      []string{"ai-agent:*", "task:task-new:stop"},
	}})

	path := "/v1/client/ai-agent/tasks/task-new/agent-assignments/agent-public-openclaw"
	aiAgentSmokeRequest(t, server, http.MethodDelete, path, token, `{}`, http.StatusNotFound)

	path = "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-new/agent-assignments/agent-public-openclaw"
	aiAgentSmokeRequest(t, server, http.MethodDelete, path, token, `{`, http.StatusBadRequest)
}

func TestHTTPAIAgentClientDeleteTaskAgentAssignmentDoesNotMutateWhenDurableCancelFails(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := aiAgentStore.AssignAIAgentTask(ctx, principal, "task-delete-cancel-fails", AssignAIAgentTaskRequest{
		AgentID:      "agent-public-openclaw",
		AssignmentID: "asn-delete-missing-durable",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, principal.PrincipalID),
	}).Handler()

	path := "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/tasks/task-delete-cancel-fails/agent-assignments/agent-public-openclaw"
	req := httptest.NewRequest(http.MethodDelete, path, strings.NewReader(`{"reason":"participant removed"}`))
	req.Header.Set("Authorization", "Bearer ai-agent-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code == http.StatusAccepted {
		t.Fatalf("delete should fail when durable cancel fails: body=%s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), "assignment asn-delete-missing-durable not found") {
		t.Fatalf("delete error = status %d body %s", resp.Code, resp.Body.String())
	}
	threads, err := aiAgentStore.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads)
	}
	thread := threads.Threads[0]
	if thread.AssignmentState != AgentAssignmentStateRunning ||
		thread.WorkStatus != AgentWorkStatusRunning ||
		!taskThreadHasActiveStream(thread) {
		t.Fatalf("failed durable delete must not pre-stop thread: %+v", thread)
	}
}
