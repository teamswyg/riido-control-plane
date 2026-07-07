package riidoaiserver

import (
	"fmt"
	"strings"
)

func (s *Store) handleReadToolApproval(
	state *storeState,
	agentID string,
	assignmentID string,
	approvalID string,
) (ToolApprovalResult, *ToolApprovalDecision, ToolApprovalRequest, bool, error) {
	key := toolApprovalKey(assignmentID, approvalID)
	approval, ok := state.toolApprovals[key]
	if !ok {
		return ToolApprovalResult{}, nil, ToolApprovalRequest{}, false, fmt.Errorf("tool approval %s not found", approvalID)
	}
	if agentID = strings.TrimSpace(agentID); agentID != "" && approval.AgentID != agentID {
		return ToolApprovalResult{}, nil, ToolApprovalRequest{}, false, fmt.Errorf("tool approval %s belongs to agent %s", approvalID, approval.AgentID)
	}
	decision, hasDecision := state.toolApprovalDecisions[key]
	if approval.Status == ApprovalPending && !approval.ExpiresAt.IsZero() && !s.now().Before(approval.ExpiresAt) {
		approval.Status = ApprovalTimedOut
		state.toolApprovals[key] = approval
		s.signalToolApprovalWaiters(state, key)
		result := toolApprovalResultFor(approval, nil, s.now())
		return result, nil, approval, true, nil
	}
	if hasDecision {
		result := toolApprovalResultFor(approval, &decision, s.now())
		return result, &decision, approval, false, nil
	}
	result := toolApprovalResultFor(approval, nil, s.now())
	return result, nil, approval, false, nil
}
