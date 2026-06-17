package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const defaultToolApprovalTTL = 5 * time.Minute

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

func (s *Store) DecideToolApproval(ctx context.Context, taskID string, decision ToolApprovalDecision) (ToolApprovalResult, *ToolApprovalDecision, error) {
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

func (s *Store) WaitForToolApproval(ctx context.Context, agentID, assignmentID, approvalID string, hold, tick time.Duration) (ToolApprovalResult, *ToolApprovalDecision, error) {
	result, decision, err := s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
	if err != nil || result.Status.IsTerminal() || hold <= 0 {
		return result, decision, err
	}
	key := toolApprovalKey(assignmentID, approvalID)
	signal, release, err := s.registerToolApprovalWaiter(ctx, key)
	if err != nil {
		return ToolApprovalResult{}, nil, err
	}
	defer release()
	result, decision, err = s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
	if err != nil || result.Status.IsTerminal() {
		return result, decision, err
	}
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
			result, decision, err = s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
			if err != nil || result.Status.IsTerminal() {
				return result, decision, err
			}
		case <-ticker.C:
			result, decision, err = s.readToolApprovalResult(ctx, agentID, assignmentID, approvalID)
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

func (s *Store) readToolApprovalResult(ctx context.Context, agentID, assignmentID, approvalID string) (ToolApprovalResult, *ToolApprovalDecision, error) {
	reply := make(chan toolApprovalDecisionResult, 1)
	if err := s.send(ctx, readToolApprovalCmd{agentID: agentID, assignmentID: assignmentID, approvalID: approvalID, reply: reply}); err != nil {
		return ToolApprovalResult{}, nil, err
	}
	select {
	case res := <-reply:
		return res.result, res.decision, res.err
	case <-ctx.Done():
		return ToolApprovalResult{}, nil, ctx.Err()
	}
}

func (s *Store) registerToolApprovalWaiter(ctx context.Context, key string) (<-chan struct{}, func(), error) {
	reply := make(chan registerToolApprovalWaiterResult, 1)
	if err := s.send(ctx, registerToolApprovalWaiterCmd{key: key, reply: reply}); err != nil {
		return nil, nil, err
	}
	var registered registerToolApprovalWaiterResult
	select {
	case registered = <-reply:
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	release := func() {
		done := make(chan struct{}, 1)
		if err := s.send(context.Background(), unregisterToolApprovalWaiterCmd{key: key, id: registered.id, reply: done}); err == nil {
			<-done
		}
	}
	return registered.ch, release, nil
}

func (s *Store) handleCreateToolApproval(state *storeState, agentID string, req ToolApprovalRequest) (ToolApprovalRequest, error) {
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

func (s *Store) handleDecideToolApproval(state *storeState, taskID string, decision ToolApprovalDecision) (ToolApprovalResult, *ToolApprovalDecision, error) {
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

func (s *Store) handleReadToolApproval(state *storeState, agentID, assignmentID, approvalID string) (ToolApprovalResult, *ToolApprovalDecision, bool, error) {
	key := toolApprovalKey(assignmentID, approvalID)
	approval, ok := state.toolApprovals[key]
	if !ok {
		return ToolApprovalResult{}, nil, false, fmt.Errorf("tool approval %s not found", approvalID)
	}
	if agentID = strings.TrimSpace(agentID); agentID != "" && approval.AgentID != agentID {
		return ToolApprovalResult{}, nil, false, fmt.Errorf("tool approval %s belongs to agent %s", approvalID, approval.AgentID)
	}
	decision, hasDecision := state.toolApprovalDecisions[key]
	if approval.Status == ApprovalPending && !approval.ExpiresAt.IsZero() && !s.now().Before(approval.ExpiresAt) {
		approval.Status = ApprovalTimedOut
		state.toolApprovals[key] = approval
		s.signalToolApprovalWaiters(state, key)
		result := toolApprovalResultFor(approval, nil, s.now())
		return result, nil, true, nil
	}
	if hasDecision {
		result := toolApprovalResultFor(approval, &decision, s.now())
		return result, &decision, false, nil
	}
	result := toolApprovalResultFor(approval, nil, s.now())
	return result, nil, false, nil
}

func (s *Store) handleRegisterToolApprovalWaiter(state *storeState, key string) (chan struct{}, int64) {
	state.nextToolApprovalWaiter++
	id := state.nextToolApprovalWaiter
	ch := make(chan struct{}, 1)
	if state.toolApprovalWaiters[key] == nil {
		state.toolApprovalWaiters[key] = map[int64]chan struct{}{}
	}
	state.toolApprovalWaiters[key][id] = ch
	return ch, id
}

func (s *Store) handleUnregisterToolApprovalWaiter(state *storeState, key string, id int64) {
	waiters := state.toolApprovalWaiters[key]
	if waiters == nil {
		return
	}
	delete(waiters, id)
	if len(waiters) == 0 {
		delete(state.toolApprovalWaiters, key)
	}
}

func (s *Store) signalToolApprovalWaiters(state *storeState, key string) {
	for _, ch := range state.toolApprovalWaiters[key] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
