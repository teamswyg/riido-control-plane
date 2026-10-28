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

type commentCompletionStore struct {
	*Store
	completion TaskEvent
}

func (s *commentCompletionStore) RecordAgentEvent(context.Context, string, AgentEventRequest) (AgentEventResponse, error) {
	return AgentEventResponse{Event: s.completion}, nil
}

type commentCallFixture struct {
	s           Server
	client      *PersistentAIAgentClientStore
	snapshots   *memoryAIAgentClientSnapshotStore
	assignments *commentCompletionStore
	evidence    verifiedCommentDelivery
	source      AIAgentTaskActionResponse
	status      int
}

func newCommentCallFixture(t *testing.T) *commentCallFixture {
	return newCommentCallFixtureWithContext(t, nil)
}

func newCommentCallFixtureWithContext(t *testing.T, contexts AssignmentContextStore) *commentCallFixture {
	t.Helper()
	f := &commentCallFixture{snapshots: &memoryAIAgentClientSnapshotStore{}, status: 200}
	f.client = openPersistentAIAgentClientStoreForDelegateTest(t, f.snapshots)
	f.evidence = verifiedCommentDelivery{EventID: "human-event", CommentID: "human-reply", RootCommentID: "root", TaskID: "handoff-task", AgentID: "agent-public-openclaw", WorkspaceID: defaultAIAgentClientWorkspaceID, PrincipalID: "user-1", Body: "Compare plans A and B for a ten-person team and produce a cost table.", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "handoff-task", ComponentType: "task", Title: "Compare plans"}, Document: AIAgentTaskContextDocument{Content: "Plan A costs 30 and B costs 50. Team size is ten."}}, ConversationMessages: []verifiedCommentContextMessage{{ID: "ordinary", Body: "Keep the human acceptance criterion", CreatedAt: time.Now().Add(-time.Minute)}}}
	key := strings.Repeat("k", 32)
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(CommentDeliveryKeyHeader) != key || r.URL.Path != "/internal/agent-comment-deliveries/human-event" {
			t.Error("unbound evidence lookup")
		}
		if f.status != 200 {
			w.WriteHeader(f.status)
			return
		}
		json.NewEncoder(w).Encode(f.evidence)
	}))
	t.Cleanup(endpoint.Close)
	f.assignments = &commentCompletionStore{Store: NewStoreWithConfig(StoreConfig{AgentRegistry: f.client})}
	t.Cleanup(f.assignments.Close)
	f.s = NewServer(ServerConfig{AIAgentClient: f.client, Assignment: f.assignments, AssignmentContext: contexts, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: endpoint.URL, Key: key}, Authorizer: assignmentHTTPAuthorizer(t, []string{"agent:agent-public-openclaw:events:write"})})
	request := httptest.NewRequest(http.MethodPost, "/internal/agent-comment-deliveries", strings.NewReader(`{"event_id":"human-event"}`))
	request.Header.Set(CommentDeliveryKeyHeader, key)
	response := httptest.NewRecorder()
	f.s.handleCommentDelivery(response, request)
	if response.Code != 202 {
		t.Fatalf("initial delivery %d %s", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &f.source); err != nil {
		t.Fatal(err)
	}
	return f
}
