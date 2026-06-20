package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreHidesLocalRuntimePathsFromClientThreadMessage(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	store.mu.Lock()
	store.taskThreads["task-local-path-copy"] = []AIAgentTaskThreadRecord{{
		ThreadID:        "thread-local-path-copy",
		TaskID:          "task-local-path-copy",
		AgentID:         "agent-owned-codex",
		RunID:           "run-local-path-copy",
		WorkStatus:      AgentWorkStatusCompleted,
		AssignmentState: AgentAssignmentStateCompleted,
		CommentKind:     AgentTaskCommentTaskCompleted,
		Message: "Go 예제는 [go/hello.go](</Users/teddy/Library/Application Support/riido/runtime/go/hello.go>)에, " +
			"Rust 예제는 [rust/hello.rs](/tmp/riido-smoke/rust/hello.rs)에 작성했습니다. " +
			"실행 산출물은 /tmp/riido-smoke/bin/hello 이고 `rustc -o /tmp/riido-smoke/bin/hello`로 검증했습니다.",
		StartedAt:   time.Now().UTC().Add(-time.Minute),
		CompletedAt: time.Now().UTC(),
		Lines:       []AgentThreadProgressLine{{Seq: 1, Message: "검증 완료 - /Users/teddy/Library/Application Support/riido/runtime/log.txt"}},
	}}
	store.mu.Unlock()
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-local-path-copy")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	message := threads.Threads[0].Message
	for _, leaked := range []string{"/Users/", "/tmp/", "file://"} {
		if strings.Contains(message, leaked) {
			t.Fatalf("thread message leaked local runtime path marker %q: %q", leaked, message)
		}
	}
	for _, want := range []string{"go/hello.go", "rust/hello.rs", "로컬 파일"} {
		if !strings.Contains(message, want) {
			t.Fatalf("thread message = %q, want to contain %q", message, want)
		}
	}
	if !strings.Contains(message, "`rustc -o 로컬 파일`") {
		t.Fatalf("thread message should preserve inline-code backticks around sanitized local path: %q", message)
	}
	if got := threads.Threads[0].Lines[0].Message; got != "검증 완료 - 로컬 파일" {
		t.Fatalf("thread progress line should hide local runtime path: %q", got)
	}
}
