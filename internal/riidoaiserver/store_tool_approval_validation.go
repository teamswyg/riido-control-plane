package riidoaiserver

import (
	"errors"
	"strings"
	"time"
)

func validateStoredToolApproval(req ToolApprovalRequest) error {
	switch {
	case strings.TrimSpace(req.ApprovalID) == "":
		return errors.New("approval_id is required")
	case strings.TrimSpace(req.AssignmentID) == "":
		return errors.New("assignment_id is required")
	case strings.TrimSpace(req.TaskID) == "":
		return errors.New("task_id is required")
	case strings.TrimSpace(req.AgentID) == "":
		return errors.New("agent_id is required")
	case strings.TrimSpace(req.ToolID) == "":
		return errors.New("tool_id is required")
	case req.Status == "":
		return errors.New("status is required")
	case req.RequestedAt.IsZero():
		return errors.New("requested_at is required")
	case req.ExpiresAt.IsZero():
		return errors.New("expires_at is required")
	case req.ExpiresAt.Before(req.RequestedAt):
		return errors.New("expires_at must be after requested_at")
	default:
		return nil
	}
}

func normalizeToolApprovalDecision(decision ToolApprovalDecision, now time.Time) ToolApprovalDecision {
	decision.ApprovalID = strings.TrimSpace(decision.ApprovalID)
	decision.AssignmentID = strings.TrimSpace(decision.AssignmentID)
	decision.DecidedBy = strings.TrimSpace(decision.DecidedBy)
	decision.Reason = strings.TrimSpace(decision.Reason)
	if decision.DecidedAt.IsZero() {
		decision.DecidedAt = now
	}
	return decision
}

func validateStoredToolApprovalDecision(decision ToolApprovalDecision) error {
	switch {
	case strings.TrimSpace(decision.ApprovalID) == "":
		return errors.New("approval_id is required")
	case strings.TrimSpace(decision.AssignmentID) == "":
		return errors.New("assignment_id is required")
	case decision.Decision != ApprovalDecisionApprove && decision.Decision != ApprovalDecisionDeny:
		return errors.New("decision must be approve or deny")
	case decision.DecidedAt.IsZero():
		return errors.New("decided_at is required")
	default:
		return nil
	}
}
