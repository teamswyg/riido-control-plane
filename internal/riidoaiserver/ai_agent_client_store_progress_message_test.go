package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreKeepsProgressMessageAcrossProviderHeartbeat(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(context.Background(), principal, "task-progress-message", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	if _, err := store.RecordAIAgentThreadProgress(context.Background(), assigned.AgentID, AgentThreadProgressBatchRequest{
		AssignmentID: "asn-progress-message",
		TaskID:       "task-progress-message",
		ThreadID:     assigned.ThreadID,
		RunID:        assigned.RunID,
		Lines: []AgentThreadProgressLine{{
			Seq:     1,
			Message: "Rust 프로젝트 생성 실행 중 - Cargo 바이너리 프로젝트를 초기화합니다.",
		}},
	}); err != nil {
		t.Fatalf("RecordAIAgentThreadProgress: %v", err)
	}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), assigned.AgentID, AgentEventRequest{}, TaskEvent{
		TaskID:       "task-progress-message",
		AssignmentID: "asn-progress-message",
		AgentID:      assigned.AgentID,
		Type:         EventProviderLog,
		State:        AssignmentRunning,
		Message:      `{"timestamp":"2026-06-04T23:14:47Z","level":"WARN","fields":{"message":"ignoring interface.icon_large"}}`,
		At:           time.Now().UTC(),
	}); err != nil {
		t.Fatalf("RecordAIAgentAssignmentEvent: %v", err)
	}
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-progress-message")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads.Threads)
	}
	want := "Rust 프로젝트 생성 실행 중 - Cargo 바이너리 프로젝트를 초기화합니다."
	if got := threads.Threads[0].Message; got != want {
		t.Fatalf("thread message = %q, want %q", got, want)
	}
}
