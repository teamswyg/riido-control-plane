package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDevelopmentAIAgentClientStoreSuppressesDuplicateAssignmentStatusFanout(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(context.Background(), principal, "task-dedupe", AssignAIAgentTaskRequest{
		AgentID: "agent-owned-codex",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}
	baseCount := countWorkStatusChangedEventsForTask(t, store, principal, "task-dedupe")
	runningEvent := TaskEvent{
		TaskID:       "task-dedupe",
		AssignmentID: "asn-dedupe",
		AgentID:      assigned.AgentID,
		Type:         EventAssignmentRunning,
		State:        AssignmentRunning,
		Message:      "provider heartbeat still running",
		At:           time.Now().UTC(),
	}
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), assigned.AgentID, AgentEventRequest{}, runningEvent); err != nil {
		t.Fatalf("first RecordAIAgentAssignmentEvent: %v", err)
	}
	afterFirstRunning := countWorkStatusChangedEventsForTask(t, store, principal, "task-dedupe")
	if afterFirstRunning != baseCount+1 {
		t.Fatalf("first running event count = %d, want %d", afterFirstRunning, baseCount+1)
	}
	runningEvent.Message = "provider heartbeat still running again"
	runningEvent.At = time.Now().UTC()
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), assigned.AgentID, AgentEventRequest{}, runningEvent); err != nil {
		t.Fatalf("second RecordAIAgentAssignmentEvent: %v", err)
	}
	afterDuplicateRunning := countWorkStatusChangedEventsForTask(t, store, principal, "task-dedupe")
	if afterDuplicateRunning != afterFirstRunning {
		t.Fatalf("duplicate running event count = %d, want unchanged %d", afterDuplicateRunning, afterFirstRunning)
	}
	completedEvent := runningEvent
	completedEvent.Type = EventAssignmentCompleted
	completedEvent.State = AssignmentCompleted
	completedEvent.Message = `<riido_log>{"code":1104,"args":{"label":"verify","summary":"passed"}}<end>작업 완료`
	completedEvent.At = time.Now().UTC()
	if err := store.RecordAIAgentAssignmentEvent(context.Background(), assigned.AgentID, AgentEventRequest{}, completedEvent); err != nil {
		t.Fatalf("completed RecordAIAgentAssignmentEvent: %v", err)
	}
	afterCompleted := countWorkStatusChangedEventsForTask(t, store, principal, "task-dedupe")
	if afterCompleted != afterFirstRunning+1 {
		t.Fatalf("completed event count = %d, want %d", afterCompleted, afterFirstRunning+1)
	}
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, "task-dedupe")
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 || threads.Threads[0].Message != "작업 완료" {
		t.Fatalf("completed thread message = %+v", threads.Threads)
	}
	if threads.Threads[0].ResultMessage != "작업 완료" {
		t.Fatalf("completed thread result_message = %+v", threads.Threads[0])
	}
	status, ok := lastWorkStatusChangedEventForTask(t, store, principal, "task-dedupe")
	if !ok || status.ResultMessage != "작업 완료" {
		t.Fatalf("completed status result_message = %+v", status)
	}
}
