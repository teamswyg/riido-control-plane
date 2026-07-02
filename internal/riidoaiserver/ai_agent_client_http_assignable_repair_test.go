package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPAssignableAgentsRepairsStaleReadModelFromAssignmentProjection(t *testing.T) {
	ctx := context.Background()
	handler, aiAgentStore, assignmentStore := newAssignableRepairServer(t)
	aiAgentSmokeRequest(
		t, handler, http.MethodPost,
		"/v1/client/ai-agent/tasks/task-assignable-repair/assignment",
		"user-token", `{"agent_id":"agent-public-openclaw"}`, http.StatusAccepted,
	)
	poll := pollAssignableRepairAssignment(t, ctx, assignmentStore)
	recordAssignableRepairCompleted(t, ctx, assignmentStore, poll.Assignment.ID)
	stale, err := aiAgentStore.ListAIAgentTaskAssignableAgents(
		ctx, AuthorizationResult{PrincipalID: "user-1"}, "task-assignable-repair",
	)
	if err != nil {
		t.Fatalf("ListAIAgentTaskAssignableAgents before repair: %v", err)
	}
	if staleAgent := agentByID(stale.Agents, "agent-public-openclaw"); staleAgent.AssignedTaskCount != 1 ||
		staleAgent.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("test setup should leave assignable stale before HTTP repair: %+v", staleAgent)
	}
	body := aiAgentSmokeRequest(
		t, handler, http.MethodGet,
		"/v1/client/ai-agent/tasks/task-assignable-repair/assignable-agents",
		"user-token", "", http.StatusOK,
	)
	var assignable AgentClientListResponse
	if err := json.Unmarshal(body, &assignable); err != nil {
		t.Fatalf("assignable json: %v", err)
	}
	repaired := agentByID(assignable.Agents, "agent-public-openclaw")
	if repaired.AssignedTaskCount != 0 ||
		repaired.Editability != AgentEditabilityEditable ||
		repaired.WorkStatus != AgentWorkStatusCompleted {
		t.Fatalf("assignable should repair stale assignment projection: %+v", repaired)
	}
}
