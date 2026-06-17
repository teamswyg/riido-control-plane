package riidoaiserver

import (
	"cmp"
	"strings"
	"time"
)

func toolApprovalKey(assignmentID, approvalID string) string {
	return strings.TrimSpace(assignmentID) + "\x00" + strings.TrimSpace(approvalID)
}

func copyToolApprovalRequest(req ToolApprovalRequest) ToolApprovalRequest {
	if len(req.Metadata) > 0 {
		req.Metadata = copyStringMap(req.Metadata)
	}
	return req
}

func toolApprovalResultFor(approval ToolApprovalRequest, decision *ToolApprovalDecision, now time.Time) ToolApprovalResult {
	resolvedAt := time.Time{}
	switch {
	case decision != nil && !decision.DecidedAt.IsZero():
		resolvedAt = decision.DecidedAt
	case approval.Status == ApprovalTimedOut && !approval.ExpiresAt.IsZero():
		resolvedAt = approval.ExpiresAt
	case approval.Status.IsTerminal():
		resolvedAt = now
	}
	return ToolApprovalResult{
		ApprovalID:   approval.ApprovalID,
		AssignmentID: approval.AssignmentID,
		Status:       approval.Status,
		ResolvedAt:   resolvedAt,
	}
}

func compareToolApprovals(a, b ToolApprovalRequest) int {
	if byTime := a.RequestedAt.Compare(b.RequestedAt); byTime != 0 {
		return byTime
	}
	if byAssignment := cmp.Compare(a.AssignmentID, b.AssignmentID); byAssignment != 0 {
		return byAssignment
	}
	return cmp.Compare(a.ApprovalID, b.ApprovalID)
}
