package riidoaiserver

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Regression: every acknowledged external message survives reopening.
func TestNativeHistoryRetains101MessagesAfterReopen(t *testing.T) {
	ctx := t.Context()
	snapshots := &memoryAIAgentClientSnapshotStore{}
	store := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	action, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "review-history-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "review-import-run", conversationID: "comment:review-root", Instruction: "next Riido request", SourceCommentID: "review-trigger"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 101; i++ {
		message := nativeHistoryMessage{SessionID: "review-native", MessageID: fmt.Sprintf("external-%03d", i), Role: "user", Body: fmt.Sprintf("external message %03d", i), Ordinal: i}
		err := store.RecordAIAgentAssignmentEvent(ctx, action.AgentID, AgentEventRequest{}, TaskEvent{TaskID: action.TaskID, AssignmentID: action.AssignmentID, AgentID: action.AgentID, Type: EventRiidoLog, State: AssignmentRunning, Metadata: nativeHistoryTestMetadata(t, message)})
		if err != nil {
			t.Fatalf("message %d not acknowledged: %v", i, err)
		}
	}
	reopened := openPersistentAIAgentClientStoreForDelegateTest(t, snapshots)
	// Reimport the beginning after restart: it must already exist, even though
	// it used to fall out of the 100-message tail.
	replay := nativeHistoryMessage{SessionID: "review-native", MessageID: "external-001", Role: "user", Body: "external message 001", Ordinal: 1}
	if err := reopened.RecordAIAgentAssignmentEvent(ctx, action.AgentID, AgentEventRequest{}, TaskEvent{TaskID: action.TaskID, AssignmentID: action.AssignmentID, AgentID: action.AgentID, Type: EventRiidoLog, State: AssignmentRunning, Metadata: nativeHistoryTestMetadata(t, replay)}); err != nil {
		t.Fatal(err)
	}
	history, err := reopened.ListAIAgentTaskThreadHistory(ctx, principal, action.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	count, triggerPresent, firstPresent := 0, false, false
	for _, thread := range history.Threads {
		for _, message := range thread.Messages {
			if strings.HasPrefix(message.MessageID, "native:") {
				count++
			}
			triggerPresent = triggerPresent || message.SourceMessageID == "review-trigger"
			firstPresent = firstPresent || strings.Contains(message.Body, "external message 001")
		}
	}
	if count != 101 || !triggerPresent || !firstPresent {
		t.Fatalf("101 native messages were acknowledged but reopened history has native=%d trigger_present=%t first_native_present=%t", count, triggerPresent, firstPresent)
	}
}

func TestConversationRetentionKeepsCompletedThreadsBeyondFifty(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	for i := range 75 {
		id := fmt.Sprintf("thread-%d", i)
		store.taskThreads["task"] = append(store.taskThreads["task"], AIAgentTaskThreadRecord{ThreadID: id, TaskID: "task", AssignmentID: fmt.Sprintf("run-%d", i), ConversationID: "comment:root", AssignmentState: AgentAssignmentState(AssignmentCompleted)})
		store.taskThreadMessages[id] = []AIAgentTaskThreadHistoryMessage{{MessageID: fmt.Sprintf("native:%d", i), Body: id}}
	}
	snapshot, err := store.snapshot(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	restored := NewDevelopmentAIAgentClientStore()
	if err := restored.restoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	if len(restored.taskThreads["task"]) != 75 || len(restored.taskThreadMessages["thread-0"]) != 1 {
		t.Fatal("completed conversation anchors were trimmed")
	}
}
