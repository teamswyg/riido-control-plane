package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeAndValidateStoredToolApprovalDecision(t *testing.T) {
	now := time.Date(2026, 7, 7, 11, 0, 0, 0, time.UTC)
	decision := normalizeToolApprovalDecision(ToolApprovalDecision{
		ApprovalID:   " approval-1 ",
		AssignmentID: " asn-1 ",
		Decision:     ApprovalDecisionApprove,
		DecidedBy:    " user-1 ",
	}, now)
	if decision.ApprovalID != "approval-1" || decision.AssignmentID != "asn-1" ||
		decision.DecidedBy != "user-1" || !decision.DecidedAt.Equal(now) {
		t.Fatalf("normalized decision = %+v", decision)
	}
	if err := validateStoredToolApprovalDecision(decision); err != nil {
		t.Fatalf("valid decision rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*ToolApprovalDecision)
		want   string
	}{
		{"approval", func(d *ToolApprovalDecision) { d.ApprovalID = " " }, "approval_id is required"},
		{"assignment", func(d *ToolApprovalDecision) { d.AssignmentID = "" }, "assignment_id is required"},
		{"decision", func(d *ToolApprovalDecision) { d.Decision = "maybe" }, "decision must be approve or deny"},
		{"decided", func(d *ToolApprovalDecision) { d.DecidedAt = time.Time{} }, "decided_at is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := decision
			tc.mutate(&candidate)
			if err := validateStoredToolApprovalDecision(candidate); err == nil ||
				!strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateStoredToolApprovalDecision error = %v, want %q", err, tc.want)
			}
		})
	}
}
