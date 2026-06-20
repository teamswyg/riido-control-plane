package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDevelopmentTaskAssignmentAndParticipantRemoval(t *testing.T) {
	server := newAIAgentClientHTTPTestServer(t, []StaticTokenCredential{{
		PrincipalID: "user-1",
		Token:       "user-token",
		Scopes:      []string{"ai-agent:*", "task:task-new:read", "task:task-new:assign", "task:task-new:comment", "task:task-new:stop"},
	}})

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw"}`))
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
	if assigned.TaskID != "task-new" ||
		assigned.AgentID != "agent-public-openclaw" ||
		assigned.AssignmentState != AgentAssignmentStateRunning ||
		assigned.CommentKind != AgentTaskCommentAssignmentStarted ||
		assigned.ThreadID == "" {
		t.Fatalf("assign response = %+v", assigned)
	}

	threadsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
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
	if len(threads.Threads) != 1 || threads.ActiveStream == nil || threads.ActiveStream.ThreadID != assigned.ThreadID {
		t.Fatalf("threads after assign = %+v", threads)
	}

	messageReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-new/threads/"+assigned.ThreadID+"/messages", strings.NewReader(`{"body":"다음 작업을 이어서 진행해 주세요.","source_message_id":"message-next-1"}`))
	messageReq.Header.Set("Authorization", "Bearer user-token")
	messageResp := httptest.NewRecorder()
	server.ServeHTTP(messageResp, messageReq)
	if messageResp.Code != http.StatusAccepted {
		t.Fatalf("thread message status=%d body=%s", messageResp.Code, messageResp.Body.String())
	}
	var message AIAgentTaskActionResponse
	if err := json.Unmarshal(messageResp.Body.Bytes(), &message); err != nil {
		t.Fatalf("thread message json: %v", err)
	}
	if message.ThreadID != assigned.ThreadID ||
		message.AssignmentState != AgentAssignmentStateRunning ||
		message.CommentKind != AgentTaskCommentRuntimeProgress {
		t.Fatalf("thread message response = %+v", message)
	}

	threadsAfterMessageReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsAfterMessageReq.Header.Set("Authorization", "Bearer user-token")
	threadsAfterMessageResp := httptest.NewRecorder()
	server.ServeHTTP(threadsAfterMessageResp, threadsAfterMessageReq)
	if threadsAfterMessageResp.Code != http.StatusOK {
		t.Fatalf("threads after message status=%d body=%s", threadsAfterMessageResp.Code, threadsAfterMessageResp.Body.String())
	}
	var threadsAfterMessage AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsAfterMessageResp.Body.Bytes(), &threadsAfterMessage); err != nil {
		t.Fatalf("threads after message json: %v", err)
	}
	if len(threadsAfterMessage.Threads) != 1 || threadsAfterMessage.Threads[0].SourceMessageID != "message-next-1" {
		t.Fatalf("threads after message = %+v", threadsAfterMessage)
	}

	unassignReq := httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/tasks/task-new/assignment", strings.NewReader(`{"agent_id":"agent-public-openclaw","reason":"removed from participants"}`))
	unassignReq.Header.Set("Authorization", "Bearer user-token")
	unassignResp := httptest.NewRecorder()
	server.ServeHTTP(unassignResp, unassignReq)
	if unassignResp.Code != http.StatusAccepted {
		t.Fatalf("unassign status=%d body=%s", unassignResp.Code, unassignResp.Body.String())
	}
	var unassigned AIAgentTaskActionResponse
	if err := json.Unmarshal(unassignResp.Body.Bytes(), &unassigned); err != nil {
		t.Fatalf("unassign json: %v", err)
	}
	if unassigned.ThreadID != assigned.ThreadID ||
		unassigned.AssignmentState != AgentAssignmentStateStopped ||
		unassigned.CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("unassign response = %+v", unassigned)
	}

	threadsAfterReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/tasks/task-new/threads", nil)
	threadsAfterReq.Header.Set("Authorization", "Bearer user-token")
	threadsAfterResp := httptest.NewRecorder()
	server.ServeHTTP(threadsAfterResp, threadsAfterReq)
	if threadsAfterResp.Code != http.StatusOK {
		t.Fatalf("threads after stop status=%d body=%s", threadsAfterResp.Code, threadsAfterResp.Body.String())
	}
	var threadsAfter AIAgentTaskThreadCollectionResponse
	if err := json.Unmarshal(threadsAfterResp.Body.Bytes(), &threadsAfter); err != nil {
		t.Fatalf("threads after stop json: %v", err)
	}
	if threadsAfter.ActiveStream != nil || len(threadsAfter.Threads) != 1 || threadsAfter.Threads[0].CommentKind != AgentTaskCommentStoppedByUserRequest {
		t.Fatalf("threads after unassign = %+v", threadsAfter)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events?replay=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer user-token")
	eventsResp := httptest.NewRecorder()
	server.ServeHTTP(eventsResp, eventsReq)
	if body := eventsResp.Body.String(); !strings.Contains(body, string(AgentTaskCommentAssignmentStarted)) || !strings.Contains(body, string(AgentTaskCommentStoppedByUserRequest)) {
		t.Fatalf("events body = %q", body)
	}
}
