package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

var ErrAIAgentMembershipChanged = errors.New("agent assignment changed; refresh before removing participant")

func validateMembershipRemovalAssignment(agentID, assignmentID string, threads []AIAgentTaskThreadRecord) error {
	agentID, assignmentID = strings.TrimSpace(agentID), strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return nil
	}
	for i := len(threads) - 1; i >= 0; i-- {
		if threads[i].AgentID != agentID {
			continue
		}
		if threads[i].AssignmentID != assignmentID {
			return ErrAIAgentMembershipChanged
		}
		return nil
	}
	return nil // The existing target resolver reports a missing assignment.
}

func (s Server) validateAIAgentMembershipRemoval(ctx context.Context, principal AuthorizationResult, taskID, agentID, assignmentID string) error {
	if strings.TrimSpace(assignmentID) == "" {
		return nil
	}
	threads, err := s.aiAgent.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return err
	}
	return validateMembershipRemovalAssignment(agentID, assignmentID, threads.Threads)
}
