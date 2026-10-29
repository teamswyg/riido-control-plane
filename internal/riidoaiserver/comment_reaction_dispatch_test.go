package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestCommentReactionDeliveryDoesNotBlockAndPreservesStateOrder(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "reaction-order", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", Instruction: "run this task", SourceCommentID: "comment", commentInvocation: &CommentInvocation{EvidenceEventID: "event"}})
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	var mu sync.Mutex
	var emojis []string
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/agent-comment-deliveries/event/reaction" || r.Header.Get(CommentDeliveryKeyHeader) != "key" {
			t.Error("wrong reaction target")
		}
		var body struct {
			Emoji string `json:"emoji"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		emojis = append(emojis, body.Emoji)
		first := len(emojis) == 1
		mu.Unlock()
		if first {
			close(entered)
			<-release
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer endpoint.Close()
	server := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: endpoint.URL, Key: "key"}})
	store.mu.Lock()
	store.taskThreads[action.TaskID][0].WorkStatus = AgentWorkStatusRunning
	store.taskThreads[action.TaskID][0].AssignmentState = "running"
	store.taskThreads[action.TaskID][0].CommentKind = ""
	store.mu.Unlock()
	began := time.Now()
	done := server.scheduleCommentStatusReactions(t.Context(), principal, action.TaskID)
	if time.Since(began) > 100*time.Millisecond {
		t.Fatal("execution blocked on reaction")
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("delivery did not start")
	}
	store.mu.Lock()
	store.taskThreads[action.TaskID][0].WorkStatus = AgentWorkStatusCompleted
	store.taskThreads[action.TaskID][0].AssignmentState = "completed"
	store.mu.Unlock()
	server.scheduleCommentStatusReactions(t.Context(), principal, action.TaskID)
	once.Do(func() { close(release) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("delivery did not finish")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(emojis) != 2 || emojis[0] != "1f440" || emojis[1] != "2705" {
		t.Fatalf("status reordered: %v", emojis)
	}
}
