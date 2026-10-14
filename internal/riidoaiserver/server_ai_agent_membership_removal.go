package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) prepareAIAgentMembershipRemoval(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID, reason string) (*taskAgentMembershipRemovalFence, error) {
	// Preserve fixture adapters without durable cancellation. Production injects
	// an AssignmentCancellationStore; its captured targets and fence must agree.
	if _, ok := s.assignment.(AssignmentCancellationStore); !ok {
		if err := s.validateAIAgentMembershipRemoval(ctx, principal, taskID, agentID, assignmentID); err != nil {
			return nil, err
		}
		return nil, nil
	}
	taskID, agentID, assignmentID = strings.TrimSpace(taskID), strings.TrimSpace(agentID), strings.TrimSpace(assignmentID)
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return nil, err
	}
	targets, found, err := assignmentActionTargetsFromThreads(taskID, agentID, assignmentID, threads.Threads)
	if err != nil {
		return nil, err
	}
	if agentID == "" && found {
		agentID = targets[0].AgentID
	}
	if err := validateMembershipRemovalAssignment(agentID, assignmentID, threads.Threads); err != nil {
		return nil, err
	}
	observation := observeTaskAgentMembershipRemoval(taskID, agentID, threads.Threads)
	// The observed ID guards membership removal; cancellation covers this member.
	targets = membershipRemovalActiveTargets(taskID, agentID, threads.Threads)
	if len(targets) > 0 {
		if _, _, err := s.cancelAIAgentAssignmentTargets(ctx, targets, reason); err != nil {
			return nil, err
		}
	}
	return observation, nil
}

// ListAIAgentTaskThreads hides queued state from clients; its empty state still
// identifies a cancellable retained execution. Terminal history is excluded.
func membershipRemovalActiveTargets(taskID, agentID string, threads []AIAgentTaskThreadRecord) []aiAgentAssignmentActionTarget {
	targets := []aiAgentAssignmentActionTarget{}
	seen := map[string]bool{}
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if thread.AgentID != agentID || thread.AssignmentID == "" || seen[thread.AssignmentID] {
			continue
		}
		if thread.AssignmentState != "" && !taskThreadHasActiveStream(thread) {
			continue
		}
		seen[thread.AssignmentID] = true
		targets = append(targets, aiAgentAssignmentActionTarget{TaskID: taskID, AgentID: agentID, AssignmentID: thread.AssignmentID})
	}
	return targets
}
