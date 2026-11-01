package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestContextChangesBaselineRejectsSiblingExpiryCycleAndUnboundedWalk(t *testing.T) {
	store := NewMemoryAssignmentContextStore()
	s := NewServer(ServerConfig{AssignmentContext: store})
	current, previous := contextChangesTestSnapshot("current"), contextChangesTestSnapshot("previous")
	current.PreviousAssignmentID, current.PreviousThreadID = previous.AssignmentID, previous.ThreadID
	save := func(item AssignmentContextSnapshot) {
		t.Helper()
		if err := store.SaveAssignmentContextSnapshot(t.Context(), item); err != nil {
			t.Fatal(err)
		}
	}
	save(previous)
	save(contextChangesTestSnapshot("sibling"))
	if _, err := s.assignmentContextBaseline(t.Context(), current, "sibling"); !errors.Is(err, errAssignmentContextBaseline) {
		t.Fatal("sibling accepted", err)
	}
	previous.PreviousAssignmentID, previous.PreviousThreadID = current.AssignmentID, current.ThreadID
	save(previous)
	save(current)
	if _, err := s.assignmentContextBaseline(t.Context(), current, "missing"); !errors.Is(err, errAssignmentContextBaseline) {
		t.Fatal("cycle accepted", err)
	}
	store.expiry = map[string]int64{previous.AssignmentID: time.Now().Add(-time.Hour).Unix()}
	store.retentionReader = &memoryContextProjection{assignment: Assignment{ID: previous.AssignmentID, State: AssignmentCompleted, UpdatedAt: time.Now().Add(-48 * time.Hour)}}
	if _, err := s.assignmentContextBaseline(t.Context(), current, previous.AssignmentID); !errors.Is(err, errAssignmentContextBaseline) {
		t.Fatal("expired parent accepted", err)
	}
	store = NewMemoryAssignmentContextStore()
	s = NewServer(ServerConfig{AssignmentContext: store})
	last := contextChangesTestSnapshot("zero")
	save(last)
	for i := 0; i < assignmentContextBaselineMaxDepth+1; i++ {
		next := contextChangesTestSnapshot(fmt.Sprint(i))
		next.PreviousAssignmentID, next.PreviousThreadID = last.AssignmentID, last.ThreadID
		save(next)
		last = next
	}
	if _, err := s.assignmentContextBaseline(t.Context(), last, "zero"); !errors.Is(err, errAssignmentContextBaseline) {
		t.Fatal("unbounded chain accepted", err)
	}
	last.CapturedAt = time.Time{}
	if _, err := s.assignmentContextBaseline(t.Context(), last, last.AssignmentID); !errors.Is(err, errAssignmentContextBaseline) {
		t.Fatal("unverified self baseline accepted", err)
	}
}

type failedContextBaselineStore struct {
	*MemoryAssignmentContextStore
	err error
}

func (s failedContextBaselineStore) LoadAssignmentContextSnapshot(context.Context, string) (AssignmentContextSnapshot, bool, error) {
	return AssignmentContextSnapshot{}, false, s.err
}

func TestContextChangesBaselinePreservesStorageFailure(t *testing.T) {
	failure := errors.New("store unavailable")
	s := NewServer(ServerConfig{AssignmentContext: failedContextBaselineStore{NewMemoryAssignmentContextStore(), failure}})
	current := contextChangesTestSnapshot("current")
	current.PreviousAssignmentID, current.PreviousThreadID = "previous", "previous-thread"
	if _, err := s.assignmentContextBaseline(t.Context(), current, "previous"); !errors.Is(err, failure) {
		t.Fatal("storage error became expiry", err)
	}
}
