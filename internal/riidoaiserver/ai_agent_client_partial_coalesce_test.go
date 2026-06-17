package riidoaiserver

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestDevelopmentAIAgentClientStoreCoalescesAssistantPartialAssignmentEvents(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-partial-event", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-partial-event",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}

	recordPartialEvent := func(message string) {
		t.Helper()
		if err := store.RecordAIAgentAssignmentEvent(ctx, assigned.AgentID, AgentEventRequest{}, TaskEvent{
			TaskID:       assigned.TaskID,
			AssignmentID: assigned.AssignmentID,
			AgentID:      assigned.AgentID,
			Type:         EventRiidoLog,
			State:        AssignmentRunning,
			Message:      message,
			Metadata: map[string]string{
				metadatakeys.ProgressMessageCode.String(): strconv.Itoa(progressmessage.AssistantPartialCode),
				metadatakeys.ProgressMessageKey.String():  progressmessage.AssistantPartialKey,
			},
			At: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("RecordAIAgentAssignmentEvent: %v", err)
		}
	}
	recordPartialEvent("The answer starts")
	recordPartialEvent("The answer starts and keeps growing")

	thread := onlyThread(t, store, principal, assigned.TaskID)
	assertAssistantPartialLine(t, thread, "The answer starts and keeps growing")
}

func TestDevelopmentAIAgentClientStoreCoalescesAssistantPartialThreadProgressBatches(t *testing.T) {
	ctx := context.Background()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1"}
	assigned, err := store.AssignAIAgentTask(ctx, principal, "task-partial-batch", AssignAIAgentTaskRequest{
		AgentID:      "agent-owned-codex",
		AssignmentID: "asn-partial-batch",
	})
	if err != nil {
		t.Fatalf("AssignAIAgentTask: %v", err)
	}

	recordPartialBatch := func(seq int, message string) {
		t.Helper()
		resp, err := store.RecordAIAgentThreadProgress(ctx, assigned.AgentID, AgentThreadProgressBatchRequest{
			AssignmentID: assigned.AssignmentID,
			TaskID:       assigned.TaskID,
			ThreadID:     assigned.ThreadID,
			RunID:        assigned.RunID,
			Lines: []AgentThreadProgressLine{{
				Seq:        seq,
				Message:    message,
				MessageKey: progressmessage.AssistantPartialKey,
			}},
		})
		if err != nil {
			t.Fatalf("RecordAIAgentThreadProgress: %v", err)
		}
		if resp.AcceptedLines != 1 {
			t.Fatalf("AcceptedLines = %d, want 1", resp.AcceptedLines)
		}
	}
	recordPartialBatch(1, "First body")
	recordPartialBatch(1, "First body replaced with a fuller body")

	thread := onlyThread(t, store, principal, assigned.TaskID)
	assertAssistantPartialLine(t, thread, "First body replaced with a fuller body")
}

func onlyThread(t *testing.T, store *DevelopmentAIAgentClientStore, principal AuthorizationResult, taskID string) AIAgentTaskThreadRecord {
	t.Helper()
	threads, err := store.ListAIAgentTaskThreads(context.Background(), principal, taskID)
	if err != nil {
		t.Fatalf("ListAIAgentTaskThreads: %v", err)
	}
	if len(threads.Threads) != 1 {
		t.Fatalf("threads = %+v", threads.Threads)
	}
	return threads.Threads[0]
}

func assertAssistantPartialLine(t *testing.T, thread AIAgentTaskThreadRecord, wantMessage string) {
	t.Helper()
	if len(thread.Lines) != 1 {
		t.Fatalf("assistant.partial lines should coalesce: %+v", thread.Lines)
	}
	line := thread.Lines[0]
	if line.Message != wantMessage || line.MessageKey != progressmessage.AssistantPartialKey {
		t.Fatalf("assistant.partial line = %+v, want message %q", line, wantMessage)
	}
	if thread.Message != wantMessage {
		t.Fatalf("thread message = %q, want %q", thread.Message, wantMessage)
	}
}
