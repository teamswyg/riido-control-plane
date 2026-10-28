package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestNativeHistoryPersistsDeduplicatesAndPrecedesTrigger(t *testing.T) {
	ctx := t.Context()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	commentTime := time.Now().UTC().Add(-time.Minute)
	first, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "history-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "import-run", conversationID: "comment:root", Instruction: "next Riido request", SourceCommentID: "reply", sourceMessageObservedAt: commentTime})
	if err != nil {
		t.Fatal(err)
	}
	message := nativeHistoryMessage{SessionID: "native", MessageID: "external-1", Role: "user", Body: "<script>unsafe</script>\nexternal", Ordinal: 5}
	record := func(s *PersistentAIAgentClientStore, run string) {
		t.Helper()
		err := s.RecordAIAgentAssignmentEvent(ctx, first.AgentID, AgentEventRequest{}, TaskEvent{TaskID: first.TaskID, AssignmentID: run, AgentID: first.AgentID, Type: EventRiidoLog, State: AssignmentRunning, Metadata: nativeHistoryTestMetadata(t, message)})
		if err != nil {
			t.Fatal(err)
		}
	}
	record(store, first.AssignmentID)
	record(store, first.AssignmentID)
	reopened := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	second, err := reopened.CreateAIAgentTaskAgentAssignment(ctx, principal, first.TaskID, AssignAIAgentTaskRequest{AgentID: first.AgentID, AssignmentID: "later-import-run", conversationID: "comment:root", parentThreadID: first.ThreadID})
	if err != nil {
		t.Fatal(err)
	}
	record(reopened, second.AssignmentID)
	history, err := reopened.ListAIAgentTaskThreadHistory(ctx, principal, first.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	var imported, trigger time.Time
	for _, row := range history.Threads {
		for _, m := range row.Messages {
			if strings.HasPrefix(m.MessageID, "native:") {
				count++
				imported = m.ObservedAt
				if m.AuthorPrincipalID != "" || m.Body != "<p>&lt;script&gt;unsafe&lt;/script&gt;<br>external</p>" {
					t.Fatalf("untrusted external user was misrepresented: %+v", m)
				}
			}
			if m.SourceMessageID == "reply" {
				trigger = m.ObservedAt
			}
		}
	}
	if count != 1 || trigger.IsZero() || !trigger.Equal(commentTime) || imported.UnixMilli() >= trigger.UnixMilli() {
		t.Fatalf("dedup/order: count=%d imported=%v trigger=%v", count, imported, trigger)
	}
	// A separate root owns a separate view even if it references the same native source.
	third, err := reopened.CreateAIAgentTaskAgentAssignment(ctx, principal, first.TaskID, AssignAIAgentTaskRequest{AgentID: first.AgentID, AssignmentID: "other-root-run", conversationID: "comment:other"})
	if err != nil {
		t.Fatal(err)
	}
	record(reopened, third.AssignmentID)
	if len(reopened.taskThreadMessages[third.ThreadID]) != 1 {
		t.Fatal("dedup crossed conversation boundary")
	}
}
