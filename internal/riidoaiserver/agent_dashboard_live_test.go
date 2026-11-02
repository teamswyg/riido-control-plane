package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type dashboardJournalProbe struct {
	scopedDashboardStore
	reads int
}

func (s *dashboardJournalProbe) LoadAgentAssignmentOperations(context.Context, string) ([]AssignmentOperationRecord, error) {
	s.reads++
	return nil, errors.New("journal unavailable")
}

func TestDashboardActiveDoesNotReadHistory(t *testing.T) {
	now := time.Now().UTC()
	state := newStoreState()
	for _, a := range []Assignment{{ID: "active", AgentID: "agent", TaskID: "a", State: AssignmentRunning, CreatedAt: now}, {ID: "done", AgentID: "agent", TaskID: "d", State: AssignmentCompleted, CreatedAt: now}} {
		applyAssignmentToState(&state, a)
	}
	db := &dashboardJournalProbe{}
	s := newStoreWithConfig(StoreConfig{OperationStore: db}, state)
	defer s.Close()
	view, err := s.LoadAgentDashboardActive(context.Background(), "agent")
	if err != nil || db.reads != 0 || len(view.Runs) != 1 || view.Runs[0].Assignment.ID != "active" {
		t.Fatalf("active=%+v err=%v history reads=%d", view, err, db.reads)
	}
}
