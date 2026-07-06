package riidoaiserver

import (
	"context"
	"strings"
)

func (s Server) cancelAIAgentAssignmentsForDelete(ctx context.Context, principal AuthorizationResult, agentID string) error {
	canceller, ok := s.assignment.(AssignmentCancellationStore)
	if !ok {
		return nil
	}
	resolver, ok := s.aiAgent.(AIAgentActiveTaskThreadsForAgentResolver)
	if !ok {
		return nil
	}
	threads, err := resolver.ActiveAIAgentTaskThreadsForAgent(ctx, principal, agentID)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, thread := range threads {
		assignmentID := strings.TrimSpace(thread.AssignmentID)
		if assignmentID == "" || seen[assignmentID] || isIntentGateAssignmentID(assignmentID) {
			continue
		}
		seen[assignmentID] = true
		if _, err := canceller.CancelAssignment(ctx, thread.TaskID, CancelAssignmentRequest{
			AgentID:      thread.AgentID,
			AssignmentID: assignmentID,
			Reason:       string(AgentTaskCommentStoppedByAgentDeleted),
		}); err != nil {
			return err
		}
	}
	return nil
}
