package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

type failingInheritedContextStore struct {
	*MemoryAssignmentContextStore
	fail bool
}

func (s *failingInheritedContextStore) SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error {
	if s.fail && snapshot.TriggerKind == AssignmentContextTriggerAgentMention {
		return errors.New("context store unavailable")
	}
	return s.MemoryAssignmentContextStore.SaveAssignmentContextSnapshot(ctx, snapshot)
}

func TestCommentHandoffRetriesContextWithoutDuplicateChild(t *testing.T) {
	contexts := &failingInheritedContextStore{MemoryAssignmentContextStore: NewMemoryAssignmentContextStore(), fail: true}
	f := newCommentCallFixtureWithContext(t, contexts)
	body := mentionHTML("agent-owned-codex") + "Compare the two plans."
	if w := f.complete(t, body); w.Code != 503 {
		t.Fatalf("inheritance failure acknowledged: %d %s", w.Code, w.Body.String())
	}
	if len(f.threads(t)) != 1 {
		t.Fatal("child published before context checkpoint")
	}
	poll, err := f.assignments.PollAgent(t.Context(), "agent-owned-codex", PollRequest{DaemonID: "daemon-dev-macbook", DeviceID: "device-dev-macbook", RuntimeID: "runtime-codex-dev"})
	if err != nil || poll.Assignment != nil {
		t.Fatalf("child executable without context: %+v %v", poll, err)
	}
	contexts.fail = false
	f.client = openPersistentAIAgentClientStoreForDelegateTest(t, f.snapshots)
	f.s.aiAgent = f.client
	if w := f.complete(t, body); w.Code != 200 {
		t.Fatalf("repair retry: %d %s", w.Code, w.Body.String())
	}
	rows := f.threads(t)
	if len(rows) != 2 {
		t.Fatalf("retry child count: %d", len(rows))
	}
	child := rows[1]
	if w := f.complete(t, body); w.Code != 200 || len(f.threads(t)) != 2 {
		t.Fatal("retry duplicated child")
	}
	if got, found, err := contexts.LoadAssignmentContextSnapshot(t.Context(), child.AssignmentID); err != nil || !found || got.TriggerID != child.SourceMessageID {
		t.Fatalf("repair failed: %+v %t %v", got, found, err)
	}
}
