package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAIAgentClientDeleteCancelsDurableAssignments(t *testing.T) {
	ctx := context.Background()
	authorizer, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{
		PrincipalID: "user-1", Token: "user-token", Scopes: []string{"ai-agent:*", "task:task-delete-durable:assign"},
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

	assignReq := httptest.NewRequest(http.MethodPost, "/v1/client/ai-agent/tasks/task-delete-durable/assignment", strings.NewReader(`{"agent_id":"agent-owned-codex"}`))
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
	pollReq := PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"}
	if poll, err := assignmentStore.PollAgent(ctx, assigned.AgentID, pollReq); err != nil || poll.Action != PollStart {
		t.Fatalf("PollAgent start = %+v err=%v", poll, err)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/client/ai-agent/agents/"+assigned.AgentID, nil)
	deleteReq.Header.Set("Authorization", "Bearer user-token")
	deleteResp := httptest.NewRecorder()
	server.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	projection, ok, err := assignmentStore.LoadAssignmentProjection(ctx, assigned.AssignmentID)
	if err != nil || !ok || projection.Assignment.State != AssignmentCancelling {
		t.Fatalf("projection after delete = %+v ok=%v err=%v", projection, ok, err)
	}
	if poll, err := assignmentStore.PollAgent(ctx, assigned.AgentID, PollRequest{}); err == nil || poll.Action == PollCancel {
		t.Fatalf("empty deleted-agent poll = %+v err=%v, want validation failure", poll, err)
	}
	if poll, err := assignmentStore.PollAgent(ctx, assigned.AgentID, pollReq); err != nil || poll.Action != PollCancel {
		t.Fatalf("PollAgent cancel = %+v err=%v", poll, err)
	}
}
