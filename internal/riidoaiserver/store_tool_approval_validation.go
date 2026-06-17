package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Store) normalizeToolApprovalRequest(state *storeState, agentID string, req ToolApprovalRequest) (ToolApprovalRequest, error) {
	now := s.now()
	req = copyToolApprovalRequest(req)
	req.ApprovalID = strings.TrimSpace(req.ApprovalID)
	req.AssignmentID = strings.TrimSpace(req.AssignmentID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	req.RuntimeID = strings.TrimSpace(req.RuntimeID)
	req.ToolID = strings.TrimSpace(req.ToolID)
	req.ToolKind = strings.TrimSpace(req.ToolKind)
	req.ToolName = strings.TrimSpace(req.ToolName)
	req.ProviderRequestID = strings.TrimSpace(req.ProviderRequestID)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.AgentID == "" {
		req.AgentID = strings.TrimSpace(agentID)
	}
	if req.Status == "" {
		req.Status = ApprovalPending
	}
	if req.Status != ApprovalPending {
		return ToolApprovalRequest{}, errors.New("tool approval status must be pending")
	}
	if req.RequestedAt.IsZero() {
		req.RequestedAt = now
	}
	if req.ExpiresAt.IsZero() {
		req.ExpiresAt = req.RequestedAt.Add(defaultToolApprovalTTL)
	}
	if err := validateStoredToolApproval(req); err != nil {
		return ToolApprovalRequest{}, err
	}
	assignment := state.assignments[req.AssignmentID]
	if assignment.ID == "" {
		return ToolApprovalRequest{}, fmt.Errorf("assignment %s not found", req.AssignmentID)
	}
	if assignment.TaskID != req.TaskID {
		return ToolApprovalRequest{}, fmt.Errorf("assignment %s belongs to task %s", req.AssignmentID, assignment.TaskID)
	}
	if assignment.AgentID != req.AgentID {
		return ToolApprovalRequest{}, fmt.Errorf("assignment %s belongs to agent %s", req.AssignmentID, assignment.AgentID)
	}
	return req, nil
}

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
