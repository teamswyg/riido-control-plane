package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPBootstrapRepairsQueuedProjectionFromEagerRunningReadModel(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-queued-repair:read", "task:task-queued-repair:assign"},
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

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-queued-repair/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
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
	if assigned.WorkStatus != AgentWorkStatusRunning || assigned.AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("test setup should create eager running read-model before daemon poll: %+v", assigned)
	}
	staleBootstrap, err := aiAgentStore.BootstrapAIAgentClient(ctx, AuthorizationResult{PrincipalID: "user-1"}, ClientKindWeb)
	if err != nil {
		t.Fatalf("BootstrapAIAgentClient before repair: %v", err)
	}
	if stale := agentByID(staleBootstrap.Agents, "agent-public-openclaw"); stale.AssignedTaskCount != 1 || stale.WorkStatus != AgentWorkStatusRunning {
		t.Fatalf("test setup should leave bootstrap eager-running before HTTP repair: %+v", stale)
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
	if repaired.AssignedTaskCount != 1 ||
		repaired.Editability != AgentEditabilityBlockedAssignedTasks ||
		repaired.WorkStatus != AgentWorkStatusQueued {
		t.Fatalf("bootstrap should repair eager running read-model to queued projection: assigned=%+v repaired=%+v", assigned, repaired)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-queued-repair/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	handler.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.ActiveStream == nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].ThreadID != assigned.ThreadID ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateQueued ||
		threads.Threads[0].WorkStatus != AgentWorkStatusQueued ||
		threads.Threads[0].CommentKind != AgentTaskCommentQueuedByBusyAgent ||
		!threads.Threads[0].CompletedAt.IsZero() {
		t.Fatalf("threads after queued projection repair = %+v", threads)
	}
}
