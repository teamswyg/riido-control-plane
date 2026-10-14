package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

func (s Server) reconcileAIAgentTaskThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string) error {
	reader, ok := s.assignment.(AssignmentProjectionReader)
	if !ok {
		return nil
	}
	reconciler, ok := s.aiAgent.(AIAgentTaskThreadProjectionReconciler)
	if !ok {
		return nil
	}
	globalKey := ""
	if strings.TrimSpace(taskID) == "" {
		key, ok := s.aiAgentGlobalReconcile.reserve(principal, time.Now())
		if !ok {
			return nil
		}
		globalKey = key
	}
	_, err := reconciler.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
	if err != nil {
		s.aiAgentGlobalReconcile.forget(globalKey)
	}
	return err
}

type aiAgentAssignmentActionTarget struct {
	TaskID            string
	AgentID           string
	AssignmentID      string
	State             AssignmentState
	durableProjection *AssignmentProjection
}

func (s Server) cancelAIAgentAssignmentBeforeAction(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID, reason string) (aiAgentAssignmentActionTarget, bool, error) {
	_, ok := s.assignment.(AssignmentCancellationStore)
	if !ok {
		return aiAgentAssignmentActionTarget{}, false, nil
	}
	targets, ok, err := s.resolveAIAgentAssignmentActionTargets(ctx, principal, taskID, agentID, assignmentID)
	if err != nil || !ok {
		return aiAgentAssignmentActionTarget{}, ok, err
	}
	return s.cancelAIAgentAssignmentTargets(ctx, targets, reason)
}

func (s Server) cancelAIAgentAssignmentTargets(ctx context.Context, targets []aiAgentAssignmentActionTarget, reason string) (aiAgentAssignmentActionTarget, bool, error) {
	canceller, ok := s.assignment.(AssignmentCancellationStore)
	if !ok || len(targets) == 0 {
		return aiAgentAssignmentActionTarget{}, false, nil
	}
	primary := targets[0]
	for _, target := range targets {
		if isIntentGateAssignmentID(target.AssignmentID) {
			continue
		}
		cancelled, err := canceller.CancelAssignment(ctx, target.TaskID, CancelAssignmentRequest{
			durableProjection: target.durableProjection,
			AgentID:           target.AgentID,
			AssignmentID:      target.AssignmentID,
			Reason:            strings.TrimSpace(reason),
		})
		if err != nil {
			return primary, true, err
		}
		if target.durableProjection != nil {
			current, found, err := s.assignment.(AssignmentProjectionReader).LoadAssignmentProjection(ctx, target.AssignmentID)
			if err != nil {
				return primary, true, err
			}
			if !found || (!isTerminal(current.Assignment.State) && current.Assignment.State != AssignmentCancelling) {
				return primary, true, ErrAIAgentMembershipChanged
			}
		}
		if target.AssignmentID == primary.AssignmentID {
			primary.State = cancelled.State
		}
	}
	return primary, true, nil
}
