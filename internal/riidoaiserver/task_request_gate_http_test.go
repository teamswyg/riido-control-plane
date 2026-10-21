package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIdentifiedAssignmentHTTPRetryKeepsFirstRequest(t *testing.T) {
	handler, assignments := newIntentGateHTTPTestServer(t)
	var first AIAgentTaskActionResponse
	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/v2/client/workspaces/workspace-dev-riid/ai-agent/tasks/task-copy/agent-assignments", strings.NewReader(`{"agent_id":"agent-owned-codex","request_id":"retry-gated-request"}`))
		request.Header.Set(aiAgentTokenHeader, "user-token")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusAccepted {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var result AIAgentTaskActionResponse
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if attempt == 0 {
			first = result
		} else if result.AssignmentID != first.AssignmentID || result.ThreadID != first.ThreadID {
			t.Fatalf("duplicate assignment: %+v %+v", first, result)
		}
	}
	poll, err := assignments.PollAgent(t.Context(), first.AgentID, PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Action != PollStart {
		t.Fatalf("request did not execute: %+v %v", poll, err)
	}
}

func TestDistinctGatedRequestsHaveDurableReceipts(t *testing.T) {
	s := NewDevelopmentAIAgentClientStore()
	p := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-dev-riid"}
	ctx := t.Context()
	a := AssignAIAgentTaskRequest{AgentID: "agent-public-openclaw", RequestID: "a", intentGateRequired: true}
	b := a
	b.RequestID = "b"
	first, err := s.CreateAIAgentTaskAgentAssignment(ctx, p, "task", a)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateAIAgentTaskAgentAssignment(ctx, p, "task", b)
	if err != nil {
		t.Fatal(err)
	}
	if first.AssignmentID == second.AssignmentID {
		t.Fatal("distinct gated request reused the previous request")
	}
	if _, err = s.UnassignAIAgentTask(ctx, p, "task", UnassignAIAgentTaskRequest{AgentID: a.AgentID}); err != nil {
		t.Fatal(err)
	}
	if _, found, err := s.FindTaskRequestReceipt(ctx, p, "task", b); err != nil || !found {
		t.Fatalf("second receipt missing: %v %v", found, err)
	}
	if _, err = s.CreateAIAgentTaskAgentAssignment(ctx, p, "task", b); err != nil {
		t.Fatal(err)
	}
	threads, err := s.ListAIAgentTaskThreads(ctx, p, "task")
	if err != nil || len(threads.AssignedAgentIDs) != 0 {
		t.Fatalf("replay restored membership: %+v %v", threads, err)
	}
}
