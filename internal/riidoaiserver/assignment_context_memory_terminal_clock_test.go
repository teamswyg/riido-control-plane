package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestMemoryContextRetentionUsesFirstTerminalEventBeforeFirstRead(t *testing.T) {
	terminalAt := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	now := terminalAt.Add(-time.Hour)
	state := newStoreState()
	applyAssignmentToState(&state, Assignment{ID: "assignment", TaskID: "task", AgentID: "agent", State: AssignmentRunning, UpdatedAt: now})
	store := newStoreWithConfig(StoreConfig{Now: func() time.Time { return now }}, state)
	t.Cleanup(store.Close)
	m := NewMemoryAssignmentContextStore()
	m.now = func() time.Time { return now }
	m.bindRetention(store)
	ctx := context.Background()
	if err := m.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	now = terminalAt
	if _, err := store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: "assignment", State: AssignmentCompleted, EventType: EventAssignmentCompleted}); err != nil {
		t.Fatal(err)
	}
	now = terminalAt.Add(time.Hour)
	if _, err := store.RecordAgentEvent(ctx, "agent", AgentEventRequest{AssignmentID: "assignment", ProviderSessionID: "late-session", EventType: EventAssignmentStateUpdated}); err != nil {
		t.Fatal(err)
	}
	projection, _, err := store.LoadAssignmentProjection(ctx, "assignment")
	if err != nil || !projection.Assignment.UpdatedAt.Equal(now) {
		t.Fatalf("late metadata projection=%+v err=%v", projection, err)
	}
	now = terminalAt.Add(24 * time.Hour)
	if _, found, err := m.LoadAssignmentContextSnapshot(ctx, "assignment"); err != nil || found {
		t.Fatalf("terminal expiry extended by metadata: found=%v err=%v", found, err)
	}
}
