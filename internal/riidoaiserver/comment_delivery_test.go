package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentDeliveryVerifiedEvidenceAndRetry(t *testing.T) {
	key := strings.Repeat("k", 32)
	evidence := verifiedCommentDelivery{EventID: "event", CommentID: "comment", TaskID: "task", AgentID: "agent-public-openclaw", WorkspaceID: "workspace-dev-riid", PrincipalID: "user-1", Body: "Please inspect the task", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task", ComponentType: "task", Title: "Test task"}}}
	evidenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(CommentDeliveryKeyHeader) != key || r.URL.Path != "/internal/agent-comment-deliveries/event" {
			t.Error("evidence request missing binding")
		}
		if err := json.NewEncoder(w).Encode(evidence); err != nil {
			t.Fatal(err)
		}
	}))
	defer evidenceServer.Close()
	clientStore := NewDevelopmentAIAgentClientStore()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: clientStore})
	defer assignments.Close()
	flakyClient := &retryableProjectionClient{DevelopmentAIAgentClientStore: clientStore, fail: true}
	handler := NewServer(ServerConfig{AIAgentClient: flakyClient, Assignment: assignments, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: evidenceServer.URL, Key: key}}).Handler()
	submit := func(secret string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/internal/agent-comment-deliveries", bytes.NewBufferString(`{"event_id":"event"}`))
		r.Header.Set(CommentDeliveryKeyHeader, secret)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if denied := submit("wrong"); denied.Code != 401 {
		t.Fatalf("missing auth status %d", denied.Code)
	}
	failed := submit(key)
	if failed.Code != http.StatusServiceUnavailable {
		t.Fatalf("transient failure must retry: %d", failed.Code)
	}
	first := submit(key)
	if first.Code != 202 {
		t.Fatalf("delivery=%d %s", first.Code, first.Body.String())
	}
	retry := submit(key)
	if retry.Code != 202 {
		t.Fatalf("retry=%d %s", retry.Code, retry.Body.String())
	}
	var a, b AIAgentTaskActionResponse
	if err := json.Unmarshal(first.Body.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(retry.Body.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if a.AssignmentID == "" || a.AssignmentID != b.AssignmentID || a.ThreadID != b.ThreadID || a.AssignmentState != AgentAssignmentStateQueued {
		t.Fatalf("request duplicated: %+v %+v", a, b)
	}
	threads, err := clientStore.ListAIAgentTaskThreads(context.Background(), AuthorizationResult{PrincipalID: "user-1", WorkspaceID: evidence.WorkspaceID}, "task")
	if err != nil || len(threads.Threads) != 1 || len(threads.AssignedAgentIDs) != 1 {
		t.Fatalf("membership: %+v %v", threads, err)
	}
	evidence.Context.Component.ID = "other-task"
	if rejected := submit(key); rejected.Code != 502 {
		t.Fatalf("mismatched evidence accepted: %d", rejected.Code)
	}
}

type retryableProjectionClient struct {
	*DevelopmentAIAgentClientStore
	fail bool
}

func (c *retryableProjectionClient) CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	if c.fail {
		c.fail = false
		return AIAgentTaskActionResponse{}, errors.New("snapshot temporarily unavailable")
	}
	return c.DevelopmentAIAgentClientStore.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, req)
}
