package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

func pendingThreadToolApproval(
	ctx context.Context,
	store AssignmentToolApprovalStore,
	taskID string,
	thread AIAgentTaskThreadRecord,
) (ToolApprovalRequest, bool, error) {
	approvals, err := store.ListTaskToolApprovals(ctx, taskID)
	if err != nil {
		return ToolApprovalRequest{}, false, err
	}
	for _, approval := range approvals {
		if toolApprovalMatchesThread(approval, thread, time.Now().UTC()) {
			return approval, true, nil
		}
	}
	return ToolApprovalRequest{}, false, nil
}

func toolApprovalMatchesThread(
	approval ToolApprovalRequest,
	thread AIAgentTaskThreadRecord,
	now time.Time,
) bool {
	if approval.Status != ApprovalPending {
		return false
	}
	if !approval.ExpiresAt.IsZero() && !now.Before(approval.ExpiresAt) {
		return false
	}
	if strings.TrimSpace(approval.AssignmentID) != strings.TrimSpace(thread.AssignmentID) {
		return false
	}
	return strings.TrimSpace(approval.AgentID) == strings.TrimSpace(thread.AgentID)
}
