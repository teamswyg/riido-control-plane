package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPBootstrapRepairsStaleReadModelFromAssignmentProjection(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-bootstrap-repair:read", "task:task-bootstrap-repair:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:events:write"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	handler := NewServer(ServerConfig{
		Assignment:    assignmentStore,
		AIAgentClient: aiAgentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-bootstrap-repair/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	handler.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if _, err := assignmentStore.RecordAgentEvent(ctx, "agent-public-openclaw", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		TaskID:       "task-bootstrap-repair",
		DaemonID:     "daemon-shared-studio",
		DeviceID:     "device-shared-studio",
		RuntimeID:    "runtime-openclaw-shared",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
		Message:      "running before terminal projection",
	}); err != nil {
		t.Fatalf("RecordAgentEvent running: %v", err)
	}
	if _, err := assignmentStore.RecordAgentEvent(ctx, "agent-public-openclaw", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		TaskID:       "task-bootstrap-repair",
		DaemonID:     "daemon-shared-studio",
		DeviceID:     "device-shared-studio",
		RuntimeID:    "runtime-openclaw-shared",
		State:        AssignmentCompleted,
		EventType:    EventAssignmentCompleted,
		Message:      "completed before bootstrap repaired the client read-model",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	staleBootstrap, err := aiAgentStore.BootstrapAIAgentClient(ctx, AuthorizationResult{PrincipalID: "user-1"}, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient before repair: %v", err)
	}
	if stale := agentByID(staleBootstrap.Agents, "agent-public-openclaw"); stale.AssignedTaskCount != 1 || stale.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("test setup should leave bootstrap stale before HTTP repair: %+v", stale)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer user-token")
	bootstrapResp := httptest.NewRecorder()
	handler.ServeHTTP(bootstrapResp, bootstrapReq)
	if bootstrapResp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapResp.Code, bootstrapResp.Body.String())
	}
	var bootstrap ClientBootstrapResponse
	if err := json.Unmarshal(bootstrapResp.Body.Bytes(), &bootstrap); err != nil {
		t.Fatalf("bootstrap json: %v", err)
	}
	repaired := agentByID(bootstrap.Agents, "agent-public-openclaw")
	if repaired.AssignedTaskCount != 0 ||
		repaired.Editability != AgentEditabilityEditable ||
		repaired.WorkStatus != AgentWorkStatusCompleted {
		t.Fatalf("bootstrap should repair stale assignment projection: assigned=%+v repaired=%+v", assigned, repaired)
	}
}
