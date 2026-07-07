package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestApplyClaimedAssignmentRejectsInvalidClaim(t *testing.T) {
	now := time.Date(2026, 7, 8, 1, 2, 3, 0, time.UTC)
	base := validClaimResult(now)
	cases := []struct {
		name string
		edit func(*AssignmentClaimResult)
		want string
	}{
		{name: "missing assignment", edit: func(c *AssignmentClaimResult) { c.Assignment.ID = "" }, want: "assignment_id is required"},
		{name: "missing operation", edit: func(c *AssignmentClaimResult) { c.Operation.OperationID = "" }, want: "operation_id is required"},
		{name: "wrong operation type", edit: func(c *AssignmentClaimResult) {
			c.Operation.OperationType = AssignmentOperationAssignTask
		}, want: "operation_type"},
		{name: "operation assignment id mismatch", edit: func(c *AssignmentClaimResult) {
			c.Operation.AssignmentID = "asn-other"
		}, want: "assignment_id mismatch"},
		{name: "operation assignment mismatch", edit: func(c *AssignmentClaimResult) {
			c.Operation.Assignment.Prompt = "changed"
		}, want: "assignment mismatch"},
		{name: "operations missing primary", edit: func(c *AssignmentClaimResult) {
			other := c.Operation
			other.OperationID = "poll-start:other"
			c.Operations = []AssignmentOperationRecord{other}
		}, want: "missing primary"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claim := base
			tc.edit(&claim)
			state := newStoreState()
			err := (&Store{}).applyClaimedAssignment(&state, claim)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want contains %q", err, tc.want)
			}
			if len(state.assignments) != 0 {
				t.Fatalf("state mutated on invalid claim: %+v", state.assignments)
			}
		})
	}
}

func TestApplyClaimedAssignmentNoopWhenNotClaimed(t *testing.T) {
	state := newStoreState()
	if err := (&Store{}).applyClaimedAssignment(&state, AssignmentClaimResult{}); err != nil {
		t.Fatalf("applyClaimedAssignment: %v", err)
	}
	if len(state.assignments) != 0 || len(state.events) != 0 {
		t.Fatalf("state mutated: %+v %+v", state.assignments, state.events)
	}
}

func validClaimResult(now time.Time) AssignmentClaimResult {
	record := validAssignmentOperationRecord(now)
	record.OperationType = AssignmentOperationPollStart
	record.Assignment.State = AssignmentLeased
	record.Events[0].Type = EventAssignmentLeased
	record.Events[0].State = AssignmentLeased
	record.OperationID = assignmentOperationID(record.OperationType, record.Assignment, record.Events)
	return AssignmentClaimResult{
		Claimed:    true,
		Assignment: record.Assignment,
		Operation:  record,
	}
}
