package riidoaiserver

import "context"

func (s *Store) CreateToolApproval(ctx context.Context, agentID string, req ToolApprovalRequest) (ToolApprovalRequest, error) {
	reply := make(chan toolApprovalCreateResult, 1)
	if err := s.send(ctx, createToolApprovalCmd{agentID: agentID, req: req, reply: reply}); err != nil {
		return ToolApprovalRequest{}, err
	}
	select {
	case res := <-reply:
		return res.approval, res.err
	case <-ctx.Done():
		return ToolApprovalRequest{}, ctx.Err()
	}
}

func (s *Store) ListTaskToolApprovals(ctx context.Context, taskID string) ([]ToolApprovalRequest, error) {
	reply := make(chan toolApprovalListResult, 1)
	if err := s.send(ctx, listToolApprovalsCmd{taskID: taskID, reply: reply}); err != nil {
		return nil, err
	}
	select {
	case res := <-reply:
		return res.approvals, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Store) DecideToolApproval(
	ctx context.Context,
	taskID string,
	decision ToolApprovalDecision,
) (ToolApprovalResult, *ToolApprovalDecision, error) {
	reply := make(chan toolApprovalDecisionResult, 1)
	if err := s.send(ctx, decideToolApprovalCmd{taskID: taskID, decision: decision, reply: reply}); err != nil {
		return ToolApprovalResult{}, nil, err
	}
	select {
	case res := <-reply:
		return res.result, res.decision, res.err
	case <-ctx.Done():
		return ToolApprovalResult{}, nil, ctx.Err()
	}
}

func (s *Store) readToolApprovalResult(
	ctx context.Context,
	agentID string,
	assignmentID string,
	approvalID string,
) (ToolApprovalResult, *ToolApprovalDecision, ToolApprovalRequest, error) {
	reply := make(chan toolApprovalDecisionResult, 1)
	cmd := readToolApprovalCmd{agentID: agentID, assignmentID: assignmentID, approvalID: approvalID, reply: reply}
	if err := s.send(ctx, cmd); err != nil {
		return ToolApprovalResult{}, nil, ToolApprovalRequest{}, err
	}
	select {
	case res := <-reply:
		return res.result, res.decision, res.approval, res.err
	case <-ctx.Done():
		return ToolApprovalResult{}, nil, ToolApprovalRequest{}, ctx.Err()
	}
}
