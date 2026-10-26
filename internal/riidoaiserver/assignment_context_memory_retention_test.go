package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryContextProjection struct{ assignment Assignment }

func (p *memoryContextProjection) LoadAssignmentProjection(context.Context, string) (AssignmentProjection, bool, error) {
	return AssignmentProjection{Assignment: p.assignment}, true, nil
}

func TestMemoryContextRetentionPrunesOnlyPayloadAfterTerminal(t *testing.T) {
	now := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	m := NewMemoryAssignmentContextStore()
	m.now = func() time.Time { return now }
	p := &memoryContextProjection{assignment: Assignment{ID: "assignment", State: AssignmentRunning, UpdatedAt: now}}
	m.retentionReader = p
	ctx := context.Background()
	if err := m.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); err != nil {
		t.Fatal(err)
	}
	if err := m.SaveAssignmentContextComments(ctx, "assignment", [][]AIAgentTaskContextComment{{{ID: "comment"}}}); err != nil {
		t.Fatal(err)
	}
	if err := m.AppendAssignmentContextReads(ctx, []AssignmentContextRead{{AssignmentID: "assignment"}}); err != nil {
		t.Fatal(err)
	}
	p.assignment.State = AssignmentFailed
	if _, found, err := m.LoadAssignmentContextSnapshot(ctx, "assignment"); err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	now = now.Add(24 * time.Hour)
	if _, found, err := m.LoadAssignmentContextCommentPage(ctx, "assignment", 0); err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if len(m.snapshots) != 0 || len(m.comments) != 0 || len(m.reads["assignment"]) != 1 {
		t.Fatal("wrong retention partition")
	}
	if err := m.SaveAssignmentContextSnapshot(ctx, AssignmentContextSnapshot{AssignmentID: "assignment"}); !errors.Is(err, errAssignmentContextExpired) {
		t.Fatalf("recapture=%v", err)
	}
}
