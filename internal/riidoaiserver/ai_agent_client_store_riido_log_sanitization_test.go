package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreStripsPersistedRiidoLogBlocksFromThreadMessage(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	store.mu.Lock()
	store.taskThreads["task-legacy-terminal-copy"] = []AIAgentTaskThreadRecord{{
		ThreadID:        "thread-legacy-terminal-copy",
		TaskID:          "task-legacy-terminal-copy",
		AgentID:         "agent-owned-codex",
		RunID:           "run-legacy-terminal-copy",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
		Message: `<riido_log>{"code":1001,"args":{}}<end>` +
			`<riido_log>{"code":1104,"args":{"label":"verify","summary":"go test passed"}}<end>` +
			"완료했습니다.\n\n- `go test ./...` 통과",
		StartedAt:   time.Now().UTC().Add(-time.Minute),
		CompletedAt: time.Now().UTC(),
		Lines: []AgentThreadProgressLine{{
			Seq:     1,
			Message: "검증 완료 - go test passed",
		}},
	}}
	store.mu.Unlock()
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-legacy-terminal-copy")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads.Threads)
	}
	message := threads.Threads[0].Message
	if strings.Contains(message, "<riido_log>") || strings.Contains(message, `"code":`) {
		t.Fatalf("thread message leaked internal riido_log transport: %q", message)
	}
	if got, want := message, "완료했습니다.\n\n- `go test ./...` 통과"; got != want {
		t.Fatalf("thread message = %q, want %q", got, want)
	}
}
