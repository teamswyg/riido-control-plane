package riidoaiserver

import (
	"slices"
	"testing"
	"time"
)

func TestCompareToolApprovalsOrdersByRequestTimeAssignmentThenApproval(t *testing.T) {
	base := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	approvals := []ToolApprovalRequest{
		{AssignmentID: "asn-b", ApprovalID: "approval-b", RequestedAt: base},
		{AssignmentID: "asn-a", ApprovalID: "approval-b", RequestedAt: base},
		{AssignmentID: "asn-a", ApprovalID: "approval-a", RequestedAt: base},
		{AssignmentID: "asn-z", ApprovalID: "approval-z", RequestedAt: base.Add(-time.Second)},
	}

	slices.SortFunc(approvals, compareToolApprovals)

	got := []string{
		approvals[0].AssignmentID + "/" + approvals[0].ApprovalID,
		approvals[1].AssignmentID + "/" + approvals[1].ApprovalID,
		approvals[2].AssignmentID + "/" + approvals[2].ApprovalID,
		approvals[3].AssignmentID + "/" + approvals[3].ApprovalID,
	}
	want := []string{
		"asn-z/approval-z",
		"asn-a/approval-a",
		"asn-a/approval-b",
		"asn-b/approval-b",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("sorted approvals = %v, want %v", got, want)
	}
}
