package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// RIID-5842: the list count mirrors what the comments tab shows.
func TestCountVisibleAgentCommentMessagesMirrorsCommentsTab(t *testing.T) {
	threads := []AIAgentTaskThreadHistoryRecord{
		{ThreadID: "comment-run", Messages: []AIAgentTaskThreadHistoryMessage{
			{MessageID: "copied-comment", Role: AIAgentTaskThreadMessageRoleUser, SourceMessageID: "riido-comment-1", Body: "monolith already counts this"},
			{MessageID: "started", Role: AIAgentTaskThreadMessageRoleAgent, CommentKind: AgentTaskCommentAssignmentStarted, Body: "started"},
			{MessageID: "progress", Role: AIAgentTaskThreadMessageRoleProgress, Body: "reading files"},
			{MessageID: "empty", Role: AIAgentTaskThreadMessageRoleAgent, Body: "  "},
			{MessageID: "answer", Role: AIAgentTaskThreadMessageRoleAgent, CommentKind: AgentTaskCommentTaskCompleted, ResultMessage: "done"},
		}},
		{ThreadID: "standalone", Messages: []AIAgentTaskThreadHistoryMessage{
			{MessageID: "typed-reply", Role: AIAgentTaskThreadMessageRoleUser, SourceMessageID: "thread-message:receipt", Body: "follow-up"},
			{MessageID: "external", Role: AIAgentTaskThreadMessageRoleUser, Body: "from Slack"},
			{MessageID: "question", Role: AIAgentTaskThreadMessageRoleAgent, CommentKind: AgentTaskCommentNeedsInput, Body: "which branch?"},
		}},
		{ThreadID: "rerun", Messages: []AIAgentTaskThreadHistoryMessage{
			{MessageID: "answer", Role: AIAgentTaskThreadMessageRoleAgent, ResultMessage: "done"},
		}},
	}
	if got := countVisibleAgentCommentMessages(threads); got != 4 {
		t.Fatalf("count = %d, want answer, typed-reply, external, question", got)
	}
}

func TestAgentCommentCountPushesTaskCountOnceAndSkipsUnchangedRepair(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(t.Context(), principal, "count-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", Instruction: "summarize this task"})
	if err != nil {
		t.Fatal(err)
	}
	history, err := store.ListAIAgentTaskThreadHistory(t.Context(), AuthorizationResult{WorkspaceID: principal.WorkspaceID, Roles: []AgentCatalogRole{AgentCatalogRoleAdmin}}, action.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	want := countVisibleAgentCommentMessages(history.Threads)
	var mu sync.Mutex
	var bodies []map[string]any
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != agentCommentCountPath || r.Header.Get(CommentDeliveryKeyHeader) != "key" {
			t.Errorf("wrong count target %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		bodies = append(bodies, body)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer endpoint.Close()
	received := func() []map[string]any {
		mu.Lock()
		defer mu.Unlock()
		return append([]map[string]any(nil), bodies...)
	}

	unconfigured := NewServer(ServerConfig{AIAgentClient: store})
	unconfigured.scheduleAgentCommentCount(t.Context(), principal.WorkspaceID, action.TaskID)

	server := NewServer(ServerConfig{AIAgentClient: store, CommentDelivery: CommentDeliveryConfig{EvidenceBaseURL: endpoint.URL, Key: "key"}})
	began := time.Now()
	server.scheduleAgentCommentCount(t.Context(), principal.WorkspaceID, action.TaskID)
	if time.Since(began) > 100*time.Millisecond {
		t.Fatal("request blocked on the count push")
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(received()) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	server.repairAgentCommentCount(t.Context(), principal.WorkspaceID, action.TaskID)
	time.Sleep(200 * time.Millisecond)

	got := received()
	if len(got) != 1 {
		t.Fatalf("pushes = %v", got)
	}
	if got[0]["workspace_id"] != principal.WorkspaceID || got[0]["task_id"] != action.TaskID || got[0]["message_count"] != float64(want) || got[0]["observed_at"] == "" {
		t.Fatalf("push = %v, want count %d", got[0], want)
	}
}
