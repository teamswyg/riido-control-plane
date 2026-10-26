package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentDeliveryCapturesAssignmentContextOnce(t *testing.T) {
	key := strings.Repeat("k", 32)
	evidence := verifiedCommentDelivery{EventID: "event", CommentID: "comment-7", TaskID: "task", AgentID: "agent-public-openclaw", WorkspaceID: "workspace-dev-riid", PrincipalID: "user-1", Body: "요약 부탁", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task", ComponentType: "task", Title: "마케팅"}, Document: AIAgentTaskContextDocument{Content: "본문"}}}
	evidenceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode(evidence) }))
	defer evidenceServer.Close()
	clientStore := NewDevelopmentAIAgentClientStore()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: clientStore})
	defer assignments.Close()
	contexts := NewMemoryAssignmentContextStore()
	flaky := &retryableProjectionClient{DevelopmentAIAgentClientStore: clientStore, fail: true}
	handler := NewServer(ServerConfig{AIAgentClient: flaky, Assignment: assignments, AssignmentContext: contexts, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: evidenceServer.URL, Key: key}}).Handler()
	var delivered AIAgentTaskActionResponse
	for range 3 { // a failed publication, the delivery, and a replayed retry
		r := httptest.NewRequest(http.MethodPost, "/internal/agent-comment-deliveries", bytes.NewBufferString(`{"event_id":"event"}`))
		r.Header.Set(CommentDeliveryKeyHeader, key)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code == http.StatusAccepted {
			_ = json.Unmarshal(w.Body.Bytes(), &delivered)
		}
	}
	snapshot, found, _ := contexts.LoadAssignmentContextSnapshot(context.Background(), delivered.AssignmentID)
	if !found || snapshot.TriggerKind != AssignmentContextTriggerComment || snapshot.TriggerID != "comment-7" || snapshot.RequesterPrincipalID != "user-1" {
		t.Fatalf("snapshot=%+v found=%v", snapshot, found)
	}
	if reads, _ := contexts.ListAssignmentContextReads(context.Background(), delivered.AssignmentID); len(reads) != 1 || reads[0].Via != AssignmentContextViaPrompt {
		t.Fatalf("retries must not duplicate prompt reads: %+v", reads)
	}
}

func TestThreadMessageCapturesAssignmentContext(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	contexts := NewMemoryAssignmentContextStore()
	auth, err := NewStaticTokenAuthorizer([]StaticTokenCredential{{PrincipalID: "user-1", Token: "native-owner-token", Scopes: []string{"riido:*"}}})
	if err != nil {
		t.Fatal(err)
	}
	f.handler = NewServer(ServerConfig{Assignment: f.store, AIAgentClient: f.client, TaskContext: &assignmentHTTPTaskContextReader{contextSnapshot: func() AIAgentTaskContext {
		c := aiAgentTaskContextHTTPFixture()
		c.Component.ID = "task-new"
		return c
	}()}, Authorizer: auth, AssignmentContext: contexts}).Handler()
	first := f.finishFirst(t, AssignmentCompleted)
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{Body: "이어서 해줘"}, http.StatusAccepted, &next)
	snapshot, found, _ := contexts.LoadAssignmentContextSnapshot(context.Background(), next.AssignmentID)
	if !found || snapshot.TriggerKind != AssignmentContextTriggerThreadMessage || snapshot.TriggerID != first.ThreadID {
		t.Fatalf("snapshot=%+v found=%v", snapshot, found)
	}
}
