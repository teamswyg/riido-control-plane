package riidoaiserver

import (
	"context"
	"sync"
	"testing"
	"time"
)

type conditionalTestOperations struct {
	mu          sync.Mutex
	projection  AssignmentProjection
	records     []AssignmentOperationRecord
	claimOnSave bool
}

func (s *conditionalTestOperations) Close() error { return nil }
func (s *conditionalTestOperations) LoadAssignmentProjection(context.Context, string) (AssignmentProjection, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.projection, true, nil
}
func (s *conditionalTestOperations) SaveAssignmentOperation(_ context.Context, record AssignmentOperationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.claimOnSave {
		s.claimOnSave = false
		s.projection.Assignment.State = AssignmentLeased
		s.projection.Assignment.LeaseToken = "winner-lease"
		s.projection.LastEventSeq++
	}
	expected := record.ExpectedProjection
	if expected == nil || expected.Assignment.State != s.projection.Assignment.State || expected.LastEventSeq != s.projection.LastEventSeq {
		return errAssignmentProjectionConflict
	}
	s.records = append(s.records, record)
	s.projection = AssignmentProjection{Assignment: record.Assignment, LastEventSeq: assignmentOperationLastEventSeq(record)}
	return nil
}
func TestDurableCancelRetriesClaimRaceAndFencesLateDaemonEvent(t *testing.T) {
	for _, claim := range []bool{false, true} {
		t.Run(map[bool]string{false: "queued", true: "claim-wins"}[claim], func(t *testing.T) {
			assignment := Assignment{ID: "assignment", TaskID: "task", AgentID: "agent", State: AssignmentQueued, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			ops := &conditionalTestOperations{projection: AssignmentProjection{Assignment: assignment, LastEventSeq: 10}, claimOnSave: claim}
			first := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer first.Close()
			second := NewStoreWithConfig(StoreConfig{OperationStore: ops})
			defer second.Close()
			request := CancelAssignmentRequest{AgentID: "agent", AssignmentID: "assignment"}
			// Neither replica starts with this assignment in its local map.
			cancelled, err := first.CancelAssignment(context.Background(), "task", request)
			want := AssignmentCancelled
			if claim {
				want = AssignmentCancelling
			}
			if err != nil || cancelled.State != want {
				t.Fatalf("cancel=%+v err=%v", cancelled, err)
			}
			replay, err := second.CancelAssignment(context.Background(), "task", request)
			if err != nil || replay.State != want || len(ops.records) != 1 {
				t.Fatalf("retry=%+v err=%v records=%d", replay, err, len(ops.records))
			}
			_, err = second.RecordAgentEvent(context.Background(), "agent", AgentEventRequest{AssignmentID: "assignment", TaskID: "task", State: AssignmentRunning})
			if err == nil || ops.projection.Assignment.State != want || len(ops.records) != 1 {
				t.Fatalf("late event revived cancellation: %v %+v", err, ops.projection)
			}
		})
	}
}

type noClaimTestOperations struct{ *conditionalTestOperations }

func (*noClaimTestOperations) ClaimNextAssignment(context.Context, string, time.Time) (AssignmentClaimResult, error) {
	return AssignmentClaimResult{}, nil
}
func TestDurableQueueNeverFallsBackToStaleLocalAssignment(t *testing.T) {
	store := NewStoreWithConfig(StoreConfig{OperationStore: &noClaimTestOperations{&conditionalTestOperations{}}})
	defer store.Close()
	state := newStoreState()
	assignment := Assignment{ID: "cancelled-elsewhere", AgentID: "agent", TaskID: "task", State: AssignmentQueued}
	applyAssignmentToState(&state, assignment)
	response, _, _, _, err := store.handlePoll(&state, "agent", PollRequest{}, true)
	if err != nil || response.Action == PollStart || response.Assignment != nil {
		t.Fatalf("executed a stale queue entry: %+v %v", response, err)
	}
}
