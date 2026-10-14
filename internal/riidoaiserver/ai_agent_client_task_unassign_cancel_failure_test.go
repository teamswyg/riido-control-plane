package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientUnassignTaskDoesNotMutateWhenDurableCancelFails(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	t.Cleanup(assignmentStore.Close)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	assigned, err := aiAgentStore.AssignAIAgentTask(ctx, principal, "task-unassign-cancel-fails", AssignAIAgentTaskRequest{
		AgentID: "agent-public-openclaw", AssignmentID: "asn-unassign-missing-durable",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		Authorizer:    aiAgentClientHTTPAuthorizer(t, []string{"ai-agent:*"}, principal.PrincipalID),
	}).Handler()
	path := "/v1/client/ai-agent/tasks/task-unassign-cancel-fails/assignment"
	resp := serveUnassignTaskBoundary(server, path, "ai-agent-token", `{"agent_id":"agent-public-openclaw","reason":"removed"}`)
	if resp.Code == 202 || !strings.Contains(resp.Body.String(), "assignment asn-unassign-missing-durable not found") {
		t.Fatalf("unassign should fail on durable cancel: status=%d body=%s", resp.Code, resp.Body.String())
	}
	threads, err := aiAgentStore.ListAIAgentTaskThreads(ctx, principal, assigned.TaskID)
	if err != nil || len(threads.Threads) != 1 {
		t.Fatalf("threads=%+v err=%v", threads, err)
	}
	thread := threads.Threads[0]
	if !taskThreadHasActiveStream(thread) ||
		thread.AssignmentState != AgentAssignmentStateRunning ||
		thread.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("failed durable unassign must not pre-stop thread: %+v", thread)
	}
}
