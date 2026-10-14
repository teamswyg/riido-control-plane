package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPAgentHeartbeatStaleLeaseClosesAIAgentTaskThreadReadModel(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 5, 2, 0, 0, 0, time.UTC)
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	operations := &runtimeFakeActiveLeaseOperationStore{}
	assignmentStore := NewStoreWithConfig(StoreConfig{
		AgentRegistry:       aiAgentStore,
		ActiveLeaseDuration: 20 * time.Second,
		Now:                 func() time.Time { return now },
		OperationStore:      operations,
	})
	defer assignmentStore.Close()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-stale:read", "task:task-stale:assign"},
	}, {
		PrincipalID: "daemon-shared-studio",
		Token:       "daemon-token",
		Scopes:      []string{"agent:*:heartbeat"},
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

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stale/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
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
	if poll.Assignment == nil || poll.Assignment.ID == "" || poll.Assignment.LeaseToken == "" {
		t.Fatalf("poll response = %+v", poll)
	}
	operations.activeFound = true
	operations.activeLease = AssignmentActiveLease{
		AgentID:            poll.Assignment.AgentID,
		ActiveAssignmentID: poll.Assignment.ID,
		LeaseToken:         poll.Assignment.LeaseToken,
		HeartbeatAt:        now,
		LeaseExpiresAt:     now.Add(20 * time.Second),
		LeaseExpiresUnixMS: now.Add(20 * time.Second).UnixMilli(),
	}
	now = now.Add(21 * time.Second)

	heartbeatBody := `{"daemon_id":"daemon-shared-studio","device_id":"device-shared-studio","runtime_id":"runtime-openclaw-shared","active_assignment_ids":["` + poll.Assignment.ID + `"],"running_task_ids":["task-stale"]}`
	heartbeatReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-public-openclaw/heartbeat", strings.NewReader(heartbeatBody))
	heartbeatReq.Header.Set("Authorization", "Bearer daemon-token")
	heartbeatResp := httptest.NewRecorder()
	handler.ServeHTTP(heartbeatResp, heartbeatReq)
	if heartbeatResp.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", heartbeatResp.Code, heartbeatResp.Body.String())
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-stale/threads", nil)
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
		threads.Threads[0].WorkStatus != AgentWorkStatusFailed ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateFailed ||
		threads.Threads[0].CommentKind != AgentTaskCommentTaskFailed ||
		!strings.Contains(threads.Threads[0].Message, "active assignment lease expired") {
		t.Fatalf("threads after stale heartbeat = %+v", threads)
	}
}
