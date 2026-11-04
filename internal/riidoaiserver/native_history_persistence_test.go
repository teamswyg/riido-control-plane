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

func TestNativeHistoryCodexAuthorIsRegisteredDeviceOwner(t *testing.T) {
	ctx := t.Context()
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-1", WorkspaceID: defaultAIAgentClientWorkspaceID}
	agent := store.agents["agent-owned-codex"]
	agent.Visibility = AgentVisibilityPublic
	store.agents[agent.AgentID] = agent
	action, err := store.CreateAIAgentTaskAgentAssignment(ctx, principal, "history-task", AssignAIAgentTaskRequest{AgentID: "agent-owned-codex", AssignmentID: "run", Instruction: "Riido reply", SourceCommentID: "reply"})
	if err != nil {
		t.Fatal(err)
	}
	teammate := AuthorizationResult{PrincipalID: "user-2", WorkspaceID: principal.WorkspaceID}
	if _, err := store.CreateAIAgentTaskThreadMessage(ctx, teammate, action.TaskID, action.ThreadID, CreateAIAgentTaskThreadMessageRequest{Body: "teammate reply"}); err != nil {
		t.Fatal(err)
	}
	for i := range store.devices {
		if store.devices[i].DeviceID == "device-dev-macbook" {
			store.devices[i].OwnerPrincipalID = "mac-owner"
		}
	}
	expected := map[string]string{"teammate reply": teammate.PrincipalID}
	for i, tc := range []struct {
		body, role, device, runtime, want string
	}{
		{"Codex reply", "user", "device-dev-macbook", "runtime-codex-dev", "mac-owner"},
		{"Agent reply", "agent", "device-dev-macbook", "runtime-codex-dev", ""},
		{"Other device", "user", "device-shared-studio", "runtime-codex-dev", ""},
		{"Other runtime", "user", "device-dev-macbook", "runtime-claude-code-dev", ""},
	} {
		body := tc.body
		if tc.role == "user" {
			body = "<p>" + body + "</p>"
		}
		expected[body] = tc.want
		message := nativeHistoryMessage{SessionID: "session", MessageID: tc.body, Body: tc.body, Role: tc.role, Ordinal: i + 1}
		req := AgentEventRequest{RuntimeID: tc.runtime, DeviceID: tc.device}
		event := TaskEvent{TaskID: action.TaskID, AssignmentID: action.AssignmentID, AgentID: action.AgentID, Type: EventRiidoLog, CreatedBy: principal.PrincipalID, Metadata: nativeHistoryTestMetadata(t, message)}
		if err := store.RecordAIAgentAssignmentEvent(ctx, action.AgentID, req, event); err != nil {
			t.Fatal(err)
		}
	}
	history, err := store.ListAIAgentTaskThreadHistory(ctx, principal, action.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, row := range history.Threads {
		for _, msg := range row.Messages {
			got[msg.Body] = msg.AuthorPrincipalID
		}
	}
	for body, want := range expected {
		if author, ok := got[body]; !ok || author != want {
			t.Errorf("%q: want author %q, got history %v", body, want, got)
		}
	}
}
