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
	TaskID       string
	AgentID      string
	AssignmentID string
	State        AssignmentState
}

func (s Server) cancelAIAgentAssignmentBeforeAction(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID, reason string) (aiAgentAssignmentActionTarget, bool, error) {
	canceller, ok := s.assignment.(AssignmentCancellationStore)
	if !ok {
		return aiAgentAssignmentActionTarget{}, false, nil
	}
	targets, ok, err := s.resolveAIAgentAssignmentActionTargets(ctx, principal, taskID, agentID, assignmentID)
	if err != nil || !ok {
		return aiAgentAssignmentActionTarget{}, ok, err
	}
	primary := targets[0]
	for _, target := range targets {
		cancelled, err := canceller.CancelAssignment(ctx, target.TaskID, CancelAssignmentRequest{
			AgentID:      target.AgentID,
			AssignmentID: target.AssignmentID,
			Reason:       strings.TrimSpace(reason),
		})
		if err != nil {
			return primary, true, err
		}
		if target.AssignmentID == primary.AssignmentID {
			primary.State = cancelled.State
		}
	}
	return primary, true, nil
}
