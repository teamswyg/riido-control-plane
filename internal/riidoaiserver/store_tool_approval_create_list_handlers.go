package riidoaiserver

import (
	"errors"
	"slices"
	"strings"
)

func (s *Store) handleCreateToolApproval(
	state *storeState,
	agentID string,
	req ToolApprovalRequest,
) (ToolApprovalRequest, error) {
	approval, err := s.normalizeToolApprovalRequest(state, agentID, req)
	if err != nil {
		return ToolApprovalRequest{}, err
	}
	key := toolApprovalKey(approval.AssignmentID, approval.ApprovalID)
	if existing, ok := state.toolApprovals[key]; ok {
		return existing, nil
	}
	state.toolApprovals[key] = approval
	s.signalToolApprovalWaiters(state, key)
	return approval, nil
}

func (s *Store) handleListTaskToolApprovals(state *storeState, taskID string) ([]ToolApprovalRequest, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("task_id is required")
	}
	approvals := make([]ToolApprovalRequest, 0)
	for _, approval := range state.toolApprovals {
		if approval.TaskID == taskID {
			approvals = append(approvals, approval)
		}
	}
	slices.SortFunc(approvals, compareToolApprovals)
	return approvals, nil
}
