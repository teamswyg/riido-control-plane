package riidoaiserver

import (
	"testing"
	"time"
)

func TestAssignmentEventThreadPrefersExactAssignmentThread(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	taskID := "task-event-thread-exact"
	store.taskThreads[taskID] = []AIAgentTaskThreadRecord{
		eventThreadRecord(taskID, "thread-active", "asn-active", AgentAssignmentStateRunning),
		eventThreadRecord(taskID, "thread-exact", "asn-exact", AgentAssignmentStateStopped),
	}

	store.mu.Lock()
	thread, ok := store.assignmentEventThreadLocked(assignmentEventInput{
		TaskID: taskID, AgentID: "agent-owned-codex", AssignmentID: "asn-exact",
	})
	store.mu.Unlock()

	if !ok || thread.ThreadID != "thread-exact" {
		t.Fatalf("thread = %+v ok=%v, want exact assignment thread", thread, ok)
	}
}

func TestAssignmentEventThreadFallsBackToActiveAgentThread(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	taskID := "task-event-thread-active"
	store.taskThreads[taskID] = []AIAgentTaskThreadRecord{
		eventThreadRecord(taskID, "thread-active", "asn-active", AgentAssignmentStateRunning),
	}

	store.mu.Lock()
	thread, ok := store.assignmentEventThreadLocked(assignmentEventInput{
		TaskID: taskID, AgentID: "agent-owned-codex", AssignmentID: "asn-missing",
	})
	store.mu.Unlock()

	if !ok || thread.ThreadID != "thread-active" {
		t.Fatalf("thread = %+v ok=%v, want active fallback thread", thread, ok)
	}
}

func TestAssignmentEventThreadCreatesRuntimeThreadWhenMissing(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	at := time.Date(2026, 7, 7, 7, 0, 0, 0, time.UTC)
	input := assignmentEventInput{
		TaskID: "task-event-thread-new", AgentID: "agent-owned-codex",
		AssignmentID: "asn-new", State: AssignmentRunning,
		Type: EventRiidoLog, Message: "runtime booting", At: at,
	}

	store.mu.Lock()
	thread, ok := store.assignmentEventThreadLocked(input)
	stored := store.taskThreads[input.TaskID]
	store.mu.Unlock()

	if ok || len(stored) != 1 {
		t.Fatalf("created ok=%v stored=%d, want new stored thread", ok, len(stored))
	}
	if thread.ThreadID != threadIDForRun(input.TaskID, input.AgentID, "run-asn-new") ||
		thread.RunID != "run-asn-new" || thread.Message != "runtime booting" {
		t.Fatalf("created thread identity/message = %+v", thread)
	}
	if thread.StartedAt != at || thread.AgentSnapshotID == "" || thread.AgentSnapshot == nil {
		t.Fatalf("created thread snapshot/time = %+v", thread)
	}
}
