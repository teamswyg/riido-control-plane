package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"
)

func (s *Store) handleDecideToolApproval(
	state *storeState,
	taskID string,
	decision ToolApprovalDecision,
) (ToolApprovalResult, *ToolApprovalDecision, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ToolApprovalResult{}, nil, errors.New("task_id is required")
	}
	decision = normalizeToolApprovalDecision(decision, s.now())
	if err := validateStoredToolApprovalDecision(decision); err != nil {
		return ToolApprovalResult{}, nil, err
	}
	key := toolApprovalKey(decision.AssignmentID, decision.ApprovalID)
	approval, ok := state.toolApprovals[key]
	if !ok {
		return ToolApprovalResult{}, nil, fmt.Errorf("tool approval %s not found", decision.ApprovalID)
	}
	if approval.TaskID != taskID {
		return ToolApprovalResult{}, nil, fmt.Errorf("tool approval %s belongs to task %s", decision.ApprovalID, approval.TaskID)
	}
	if existing, ok := state.toolApprovalDecisions[key]; ok {
		result := toolApprovalResultFor(approval, &existing, s.now())
		return result, &existing, nil
	}
	if approval.Status == ApprovalPending && !approval.ExpiresAt.IsZero() && !s.now().Before(approval.ExpiresAt) {
		approval.Status = ApprovalTimedOut
		state.toolApprovals[key] = approval
		s.signalToolApprovalWaiters(state, key)
		return toolApprovalResultFor(approval, nil, s.now()), nil, nil
	}
	if approval.Status.IsTerminal() {
		result := toolApprovalResultFor(approval, nil, s.now())
		return result, nil, nil
	}
	if decision.Decision == ApprovalDecisionApprove {
		approval.Status = ApprovalApproved
	} else {
		approval.Status = ApprovalDenied
	}
	state.toolApprovals[key] = approval
	state.toolApprovalDecisions[key] = decision
	s.signalToolApprovalWaiters(state, key)
	result := toolApprovalResultFor(approval, &decision, s.now())
	return result, &decision, nil
}
