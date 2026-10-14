package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientStopCancelsDurableAssignmentForDaemonPoll(t *testing.T) {
	ctx := context.Background()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-stop-durable:read", "task:task-stop-durable:assign", "task:task-stop-durable:stop"},
	}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuthorizer: %v", err)
	}
	aiAgentStore := NewDevelopmentAIAgentClientStore()
	assignmentStore := NewStoreWithConfig(StoreConfig{AgentRegistry: aiAgentStore})
	defer assignmentStore.Close()
	server := NewServer(ServerConfig{
		AIAgentClient: aiAgentStore,
		Assignment:    assignmentStore,
		TaskContext:   &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()},
		Authorizer:    authorizer,
	}).Handler()

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stop-durable/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
	assignReq.Header.Set("Authorization", "Bearer user-token")
	assignResp := httptest.NewRecorder()
	server.ServeHTTP(assignResp, assignReq)
	if assignResp.Code != http.StatusAccepted {
		t.Fatalf("assign status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assigned AIAgentTaskActionResponse
	if err := json.Unmarshal(assignResp.Body.Bytes(), &assigned); err != nil {
		t.Fatalf("assign json: %v", err)
	}
	if assigned.AssignmentID == "" || assigned.AgentID != "agent-public-openclaw" {
		t.Fatalf("assign response = %+v", assigned)
	}
	openClawPollRequest := PollRequest{
		DaemonID:  "daemon-shared-studio",
		DeviceID:  "device-shared-studio",
		RuntimeID: "runtime-openclaw-shared",
	}
	pollStart, err := assignmentStore.PollAgent(ctx, assigned.AgentID, openClawPollRequest)
	if err != nil {
		t.Fatalf("PollAgent start: %v", err)
	}
	if pollStart.Action != PollStart || pollStart.Assignment == nil || pollStart.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll start = %+v, assigned=%+v", pollStart, assigned)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-stop-durable/stop", strings.NewReader(`{"agent_id":"agent-public-openclaw","reason":"user requested stop"}`))
	stopReq.Header.Set("Authorization", "Bearer user-token")
	stopResp := httptest.NewRecorder()
	server.ServeHTTP(stopResp, stopReq)
	if stopResp.Code != http.StatusAccepted {
		t.Fatalf("stop status=%d body=%s", stopResp.Code, stopResp.Body.String())
	}
	var stopped AIAgentTaskActionResponse
	if err := json.Unmarshal(stopResp.Body.Bytes(), &stopped); err != nil {
		t.Fatalf("stop json: %v", err)
	}
	if stopped.AssignmentID != assigned.AssignmentID ||
		stopped.AssignmentState != AgentAssignmentStateStopped ||
		stopped.WorkStatus != AgentWorkStatusIdle ||
		stopped.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("stop response = %+v, assigned=%+v", stopped, assigned)
	}
	if stopped.ActiveStream != nil {
		t.Fatalf("stop response must close ui active stream: %+v", stopped.ActiveStream)
	}
	projection, ok, err := assignmentStore.LoadAssignmentProjection(ctx, assigned.AssignmentID)
	if err != nil {
		t.Fatalf("LoadAssignmentProjection: %v", err)
	}
	if !ok || projection.Assignment.State != AssignmentCancelling {
		t.Fatalf("projection after stop = %+v ok=%v", projection, ok)
	}
	pollCancel, err := assignmentStore.PollAgent(ctx, assigned.AgentID, openClawPollRequest)
	if err != nil {
		t.Fatalf("PollAgent cancel: %v", err)
	}
	if pollCancel.Action != PollCancel || pollCancel.Assignment == nil || pollCancel.Assignment.ID != assigned.AssignmentID {
		t.Fatalf("poll cancel = %+v", pollCancel)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-stop-durable/threads", nil)
	threadsReq.Header.Set("Authorization", "Bearer user-token")
	threadsResp := httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	var threads AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads json: %v", err)
	}
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateStopped ||
		threads.Threads[0].WorkStatus != AgentWorkStatusIdle {
		t.Fatalf("threads after stop = %+v", threads)
	}

	cancelledEvent, err := assignmentStore.RecordAgentEvent(ctx, assigned.AgentID, AgentEventRequest{
		AssignmentID: assigned.AssignmentID,
		TaskID:       assigned.TaskID,
		DaemonID:     openClawPollRequest.DaemonID,
		DeviceID:     openClawPollRequest.DeviceID,
		RuntimeID:    openClawPollRequest.RuntimeID,
		State:        AssignmentCancelled,
		EventType:    EventAssignmentCancelled,
		Message:      "provider cancelled after client stop",
	})
	if err != nil {
		t.Fatalf("RecordAgentEvent cancel: %v", err)
	}
	if err := aiAgentStore.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, cancelledEvent.Event); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent cancel: %v", err)
	}
	threadsResp = httptest.NewRecorder()
	server.ServeHTTP(threadsResp, threadsReq)
	if threadsResp.Code != http.StatusOK {
		t.Fatalf("threads after cancel status=%d body=%s", threadsResp.Code, threadsResp.Body.String())
	}
	threads = AIAgentTaskThreadCollectionResponse{}
	if err := json.Unmarshal(threadsResp.Body.Bytes(), &threads); err != nil {
		t.Fatalf("threads after cancel json: %v", err)
	}
	if threads.ActiveStream != nil ||
		len(threads.Threads) != 1 ||
		threads.Threads[0].AssignmentID != assigned.AssignmentID ||
		threads.Threads[0].AssignmentState != AgentAssignmentStateStopped {
		t.Fatalf("threads after cancel = %+v", threads)
	}
}
