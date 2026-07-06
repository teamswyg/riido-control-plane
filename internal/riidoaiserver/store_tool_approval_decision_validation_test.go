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

	decision.Decision = "maybe"
	if err := validateStoredToolApprovalDecision(decision); err == nil ||
		!strings.Contains(err.Error(), "decision must be approve or deny") {
		t.Fatalf("invalid decision error = %v", err)
	}
}
