package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) resolveAIAgentAssignmentActionTargets(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID string) ([]aiAgentAssignmentActionTarget, bool, error) {
	taskID = strings.TrimSpace(taskID)
	agentID = strings.TrimSpace(agentID)
	assignmentID = strings.TrimSpace(assignmentID)
	if taskID == "" || s.aiAgent == nil {
		return nil, false, nil
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return nil, false, err
	}
	return assignmentActionTargetsFromThreads(taskID, agentID, assignmentID, threads.Threads)
}

func assignmentActionTargetsFromThreads(taskID, agentID, assignmentID string, threads []AIAgentTaskThreadRecord) ([]aiAgentAssignmentActionTarget, bool, error) {
	if assignmentID != "" {
		return assignmentActionTargetsByAssignmentID(taskID, agentID, assignmentID, threads)
	}
	if agentID != "" {
		return assignmentActionTargetsByAgentID(taskID, agentID, threads)
	}
	target, ok, err := activeAssignmentActionTarget(taskID, threads)
	if !ok || err != nil {
		return nil, ok, err
	}
	return []aiAgentAssignmentActionTarget{target}, true, nil
}

func assignmentActionTargetsByAssignmentID(taskID, agentID, assignmentID string, threads []AIAgentTaskThreadRecord) ([]aiAgentAssignmentActionTarget, bool, error) {
	target, ok, err := assignmentActionTargetByAssignmentID(taskID, agentID, assignmentID, threads)
	if !ok || err != nil {
		return nil, ok, err
	}
	return []aiAgentAssignmentActionTarget{target}, true, nil
}

func assignmentActionTargetsByAgentID(taskID, agentID string, threads []AIAgentTaskThreadRecord) ([]aiAgentAssignmentActionTarget, bool, error) {
	if targets := activeAssignmentActionTargetsForAgent(taskID, agentID, threads); len(targets) > 0 {
		return targets, true, nil
	}
	target, ok := actionTargetFromThread(taskID, agentID, threads, false)
	if !ok {
		return nil, false, nil
	}
	return []aiAgentAssignmentActionTarget{target}, true, nil
}
