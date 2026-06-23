package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestDevelopmentAIAgentClientStoreStripsPartialRiidoLogFragmentsFromProgressLines(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	store.mu.Lock()
	store.taskThreads["task-partial-riido-log"] = []AIAgentTaskThreadRecord{{
		ThreadID:        "thread-partial-riido-log",
		TaskID:          "task-partial-riido-log",
		AgentID:         "agent-owned-codex",
		RunID:           "run-partial-riido-log",
		WorkStatus:      AgentWorkStatusRunning,
		AssignmentState: AgentAssignmentStateRunning,
		Message:         "생각 중...\n<ri",
		Lines: []AgentThreadProgressLine{
			{Seq: 1, Message: "생각 중...\n<ri"},
			{Seq: 2, Message: `<riido_log>{"code":1001,"args"`},
			{Seq: 3, Message: "코드를 작성 중입니다."},
		},
	}}
	store.mu.Unlock()
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-partial-riido-log")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	assertPartialRiidoLogSanitizedThread(t, threads.Threads[0])
}

func assertPartialRiidoLogSanitizedThread(t *testing.T, thread AIAgentTaskThreadRecord) {
	t.Helper()
	if thread.Message != "생각 중..." {
		t.Fatalf("thread message = %q", thread.Message)
	}
	if len(thread.Lines) != 2 {
		t.Fatalf("lines = %+v", thread.Lines)
	}
	for _, line := range thread.Lines {
		if strings.Contains(line.Message, "<ri") || strings.Contains(line.Message, `"code":`) {
			t.Fatalf("line leaked internal riido_log transport: %+v", line)
		}
	}
}
