package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreClosesInlineCodeAfterSanitizedLocalFile(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	store.mu.Lock()
	store.taskThreads["task-inline-path-copy"] = []AIAgentTaskThreadRecord{{
		ThreadID:        "thread-inline-path-copy",
		TaskID:          "task-inline-path-copy",
		AgentID:         "agent-owned-codex",
		RunID:           "run-inline-path-copy",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
		Message: "Go 예제는 go/hello.go에, Rust 예제는 rust/hello.rs에 추가했고, " +
			"`go run ./go/hello.go`와 `rustc ./rust/hello.rs -o /tmp/riido-smoke/rust_hello && /tmp/riido-smoke/rust_hello 짧게 검증해 각각 " +
			"`Hello, World from Go!`, `Hello, World from Rust!` 출력을 확인했습니다.",
		StartedAt:   time.Now().UTC().Add(-time.Minute),
		CompletedAt: time.Now().UTC(),
	}}
	store.mu.Unlock()
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-inline-path-copy")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	message := threads.Threads[0].Message
	for _, leaked := range []string{"/Users/", "/tmp/", "file://"} {
		if strings.Contains(message, leaked) {
			t.Fatalf("thread message leaked local runtime path marker %q: %q", leaked, message)
		}
	}
	if strings.Count(message, "`")%2 != 0 {
		t.Fatalf("thread message should have balanced inline-code backticks: %q", message)
	}
	if !strings.Contains(message, "`rustc ./rust/hello.rs -o 로컬 파일 && 로컬 파일` 짧게 검증해 각각") {
		t.Fatalf("thread message should close inline code before Korean prose: %q", message)
	}
}
