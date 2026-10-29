package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestCommentStatusReactionFollowsRunAndNeverBlocksDelivery(t *testing.T) {
	key := strings.Repeat("k", 32)
	evidence := verifiedCommentDelivery{EventID: "event", CommentID: "comment", TaskID: "task", AgentID: "agent-public-openclaw", WorkspaceID: "workspace-dev-riid", PrincipalID: "user-1", Body: "검토 부탁", Context: AIAgentTaskContext{Component: AIAgentTaskContextComponent{ID: "task", ComponentType: "task", Title: "검토"}}}
	var mu sync.Mutex
	var puts []string
	failNext := 1
	monolith := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(evidence)
			return
		}
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		defer mu.Unlock()
		if r.Header.Get(CommentDeliveryKeyHeader) != key || r.URL.Path != "/internal/agent-comment-deliveries/event/reaction" {
			t.Errorf("reaction request missing binding: %s", r.URL.Path)
		}
		if failNext > 0 {
			failNext--
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		puts = append(puts, string(body))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer monolith.Close()
	clientStore := NewDevelopmentAIAgentClientStore()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: clientStore})
	defer assignments.Close()
	server := NewServer(ServerConfig{AIAgentClient: clientStore, Assignment: assignments, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: monolith.URL, Key: key}})
	deliver := func() {
		r := httptest.NewRequest(http.MethodPost, "/internal/agent-comment-deliveries", bytes.NewBufferString(`{"event_id":"event"}`))
		r.Header.Set(CommentDeliveryKeyHeader, key)
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		if w.Code != http.StatusAccepted {
			t.Fatalf("delivery = %d %s (a reaction failure must not fail delivery)", w.Code, w.Body.String())
		}
		server.commentReactions.wg.Wait()
	}
	deliver()
	deliver()
	thread := clientStore.taskThreads["task"][0].ThreadID
	setReactionThreadState(clientStore, "task", thread, AgentAssignmentStateCompleted)
	server.reconcileCommentStatusReactions("task")
	server.commentReactions.wg.Wait()
	server.reconcileCommentStatusReactions("task")
	server.commentReactions.wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	want := []string{`{"emoji":"1f440"}`, `{"emoji":"2705"}`}
	if strings.Join(puts, " ") != strings.Join(want, " ") {
		t.Fatalf("reaction writes = %v, want %v (failed write retried once, no duplicates)", puts, want)
	}
}

// acceptCommentStatusReaction lets evidence fakes ignore the status reaction
// writes that follow an accepted delivery.
func acceptCommentStatusReaction(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPut || !strings.HasSuffix(r.URL.Path, "/reaction") {
		return false
	}
	w.WriteHeader(http.StatusNoContent)
	return true
}
