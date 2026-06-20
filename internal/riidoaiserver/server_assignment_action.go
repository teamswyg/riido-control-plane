package riidoaiserver

import (
	"context"
	"strings"
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
	_, err := reconciler.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
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
	target, ok, err := s.resolveAIAgentAssignmentActionTarget(ctx, principal, taskID, agentID, assignmentID)
	if err != nil || !ok {
		return target, ok, err
	}
	cancelled, err := canceller.CancelAssignment(ctx, target.TaskID, CancelAssignmentRequest{
		AgentID:      target.AgentID,
		AssignmentID: target.AssignmentID,
		Reason:       strings.TrimSpace(reason),
	})
	target.State = cancelled.State
	return target, true, err
}

func (s Server) resolveAIAgentAssignmentActionTarget(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID string) (aiAgentAssignmentActionTarget, bool, error) {
	taskID = strings.TrimSpace(taskID)
	agentID = strings.TrimSpace(agentID)
	assignmentID = strings.TrimSpace(assignmentID)
	if taskID == "" || s.aiAgent == nil {
		return aiAgentAssignmentActionTarget{}, false, nil
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return aiAgentAssignmentActionTarget{}, false, err
	}
	if assignmentID != "" {
		return assignmentActionTargetByAssignmentID(taskID, agentID, assignmentID, threads.Threads)
	}
	if agentID != "" {
		if target, ok := actionTargetFromThread(taskID, agentID, threads.Threads, true); ok {
			return target, true, nil
		}
		target, ok := actionTargetFromThread(taskID, agentID, threads.Threads, false)
		return target, ok, nil
	}
	return activeAssignmentActionTarget(taskID, threads.Threads)
}
