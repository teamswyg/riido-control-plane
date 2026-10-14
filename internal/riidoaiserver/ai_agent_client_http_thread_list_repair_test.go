package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPTaskThreadListRepairsStaleReadModelFromAssignmentProjection(t *testing.T) {
	ctx := context.Background()
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-repair:read", "task:task-repair:assign"},
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

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-repair/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
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
	if assigned.AssignmentID == "" {
		t.Fatalf("assigned response must expose assignment_id: %+v", assigned)
	}
	poll, err := assignmentStore.PollAgent(ctx, "agent-public-openclaw", PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	})
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Assignment == nil || poll.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll response = %+v, assigned=%+v", poll, assigned)
	}

	if _, err := assignmentStore.RecordAgentEvent(ctx, "agent-public-openclaw", AgentEventRequest{
		AssignmentID: poll.Assignment.ID,
		TaskID:       "task-repair",
		DaemonID:     "daemon-shared-studio",
		DeviceID:     "device-shared-studio",
		RuntimeID:    "runtime-openclaw-shared",
		State:        AssignmentFailed,
		EventType:    EventAssignmentFailed,
		Message:      "provider process exited before client read-model update",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}
	beforeRepair, err := aiAgentStore.ListAIAgentTaskThreads(ctx, AuthorizationResult{PrincipalID: "user-1"}, "task-repair")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads before repair: %v", err)
	}
	if beforeRepair.ActiveStream == nil || beforeRepair.Threads[0].AssignmentState != AgentAssignmentStateRunning {
		t.Fatalf("test setup should leave stale active client read-model: %+v", beforeRepair)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-repair/threads", nil)
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
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].ThreadID != assigned.ThreadID ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateFailed ||
		threads.Threads[0].CommentKind != AgentTaskCommentTaskFailed ||
		threads.Threads[0].CompletedAt.IsZero() {
		t.Fatalf("threads after projection repair = %+v", threads)
	}
}
