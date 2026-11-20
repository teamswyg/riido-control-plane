package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishAgentCommentUsesCommittedReplyAndStableReplayIdentity(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.agents["agent"] = AgentClientRecord{AgentID: "agent", Name: "Reviewer", WorkspaceID: "workspace", OwnerPrincipalID: "requester", Visibility: AgentVisibilityPublic}
	store.taskThreads["task"] = []AIAgentTaskThreadRecord{{
		TaskID: "task", AgentID: "agent", ThreadID: "thread", ConversationID: "comment:root",
		AssignmentID: "assignment", RunID: "run", AssignmentState: AgentAssignmentStateCompleted,
		CommentKind: AgentTaskCommentTaskCompleted, ResultMessage: "Verified result", CompletedAt: time.Now(),
	}}
	posts := make(chan agentCommentPublication, 3)
	var attempts atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/agent-comment-publications" || r.Method != http.MethodPost || r.Header.Get(CommentDeliveryKeyHeader) != "fixture-key" {
			t.Error("unexpected publication request")
		}
		var body agentCommentPublication
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		posts <- body
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()
	s := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{
		PublishComments: true, EvidenceBaseURL: upstream.URL, Key: "fixture-key",
	}})
	event := TaskEvent{WorkspaceID: "workspace", TaskID: "task", AgentID: "agent", AssignmentID: "assignment", CreatedBy: "requester", Type: EventRiidoLog}
	if err := s.publishAgentComment(t.Context(), event); err != nil || len(posts) != 0 {
		t.Fatalf("runtime log was published: %v", err)
	}
	event.Type, event.State = EventAssignmentCompleted, AssignmentCompleted
	if err := s.publishAgentComment(t.Context(), event); err == nil {
		t.Fatal("retryable handoff failure was lost")
	}
	if err := s.publishAgentComment(t.Context(), event); err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("publication attempts = %d, want 2", len(posts))
	}
	post, replay := <-posts, <-posts
	if !reflect.DeepEqual(post, replay) {
		t.Fatalf("reply identity changed on retry: %+v vs %+v", post, replay)
	}
	if post.MessageID == "" || post.RootCommentID != "root" || post.ThreadID != "thread" || post.Body != "Verified result" || !post.Public {
		t.Fatalf("incorrect published reply: %+v", post)
	}
	agent := store.agents["agent"]
	agent.Visibility = AgentVisibilityPrivate
	store.agents["agent"] = agent
	if err := s.publishAgentComment(t.Context(), event); err != nil || len(posts) != 1 || (<-posts).Public {
		t.Fatalf("private Agent reply escaped its audience: %v", err)
	}
}
