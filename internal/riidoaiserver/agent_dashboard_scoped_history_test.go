package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

type scopedDashboardStore struct{ records []AssignmentOperationRecord }

func (*scopedDashboardStore) SaveAssignmentOperation(context.Context, AssignmentOperationRecord) error {
	return nil
}
func (*scopedDashboardStore) Close() error { return nil }
func (*scopedDashboardStore) LoadAssignmentOperations(context.Context) ([]AssignmentOperationRecord, error) {
	panic("dashboard must not replay global journal")
}

func (s *scopedDashboardStore) LoadAgentAssignmentOperations(_ context.Context, id string) ([]AssignmentOperationRecord, error) {
	return s.records, nil
}

func TestAgentDashboardUsesScopedHistoryAndRetainsStartupHistory(t *testing.T) {
	now := time.Now().UTC()
	state := newStoreState()
	applyAssignmentToState(&state, Assignment{ID: "legacy", TaskID: "old", AgentID: "agent", State: AssignmentCompleted, CreatedAt: now, UpdatedAt: now})
	applyAssignmentToState(&state, Assignment{ID: "private", TaskID: "hidden", AgentID: "other", State: AssignmentCompleted, CreatedAt: now, UpdatedAt: now})
	record := replayOperationRecord("new", "fresh", "agent", AssignmentCompleted, 1, now)
	store := newStoreWithConfig(StoreConfig{OperationStore: &scopedDashboardStore{records: []AssignmentOperationRecord{record}}}, state)
	defer store.Close()
	history, err := store.LoadAgentDashboardHistory(context.Background(), "agent")
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Runs) != 2 {
		t.Fatalf("startup or cross-replica history lost: %+v", history)
	}
	for _, run := range history.Runs {
		if run.Assignment.AgentID != "agent" {
			t.Fatal("foreign agent leaked")
		}
	}
}
