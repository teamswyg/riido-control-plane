package riidoaiserver

import (
	"context"
	"time"
)

func (s *Store) WaitForToolApproval(ctx context.Context, agentID, assignmentID, approvalID string, hold, tick time.Duration) (ToolApprovalResult, *ToolApprovalDecision, error) {
	result, decision, _, err := s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
	if err != nil || result.Status.IsTerminal() || hold <= 0 {
		return result, decision, err
	}
	key := toolApprovalKey(assignmentID, approvalID)
	signal, release, err := s.registerToolApprovalWaiter(ctx, key)
	if err != nil {
		return ToolApprovalResult{}, nil, err
	}
	defer release()
	result, decision, approval, err := s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
	if err != nil || result.Status.IsTerminal() {
		return result, decision, err
	}
	if !toolApprovalExpiresDuringHold(approval, s.now(), hold) {
		return s.waitForToolApprovalSignalOnly(ctx, agentID, assignmentID, approvalID, signal, hold, result, decision)
	}
	return s.waitForToolApprovalWithTick(ctx, agentID, assignmentID, approvalID, signal, hold, tick, result, decision)
}

func toolApprovalExpiresDuringHold(approval ToolApprovalRequest, now time.Time, hold time.Duration) bool {
	return !approval.ExpiresAt.IsZero() && !approval.ExpiresAt.After(now.Add(hold))
}

func (s *Store) waitForToolApprovalSignalOnly(ctx context.Context, agentID, assignmentID, approvalID string, signal <-chan struct{}, hold time.Duration, result ToolApprovalResult, decision *ToolApprovalDecision) (ToolApprovalResult, *ToolApprovalDecision, error) {
	deadline := time.NewTimer(hold)
	defer deadline.Stop()
	for {
		select {
		case <-signal:
			var err error
			result, decision, _, err = s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
			if err != nil || result.Status.IsTerminal() {
				return result, decision, err
			}
		case <-deadline.C:
			return result, decision, nil
		case <-ctx.Done():
			return ToolApprovalResult{}, nil, ctx.Err()
		}
	}
}

func (s *Store) waitForToolApprovalWithTick(ctx context.Context, agentID, assignmentID, approvalID string, signal <-chan struct{}, hold, tick time.Duration, result ToolApprovalResult, decision *ToolApprovalDecision) (ToolApprovalResult, *ToolApprovalDecision, error) {
	if tick <= 0 || tick > hold {
		tick = hold
	}
	deadline := time.NewTimer(hold)
	defer deadline.Stop()
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-signal:
		case <-ticker.C:
		case <-deadline.C:
			return result, decision, nil
		case <-ctx.Done():
			return ToolApprovalResult{}, nil, ctx.Err()
		}
		var err error
		result, decision, _, err = s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
		if err != nil || result.Status.IsTerminal() {
			return result, decision, err
		}
	}
}
