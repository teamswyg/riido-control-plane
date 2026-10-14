package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const defaultToolApprovalTTL = 5 * time.Minute

func (s *Store) normalizeToolApprovalRequest(
	state *storeState,
	agentID string,
	req ToolApprovalRequest,
) (ToolApprovalRequest, error) {
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
