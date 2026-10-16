package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *PersistentAIAgentClientStore) CompleteAIAgentOnboardingTask(ctx context.Context, principal AuthorizationResult, taskID string, req CompleteAIAgentOnboardingTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	// The membership key must be the normalized ID, the same one the development
	// store uses, or a padded agent_id would be recorded under a different key.
	taskID = strings.TrimSpace(taskID)
	req.AgentID = strings.TrimSpace(req.AgentID)
	// Not a publication: the membership carries no durable generation, so removal
	// fences treat it as a plain, non-executing participant.
	var pending []ClientStreamEvent
	var mutationErr error
	response, err := s.mutateTaskAgentMembership(ctx, taskID, req.AgentID, "", nil, false, func() (AIAgentTaskActionResponse, error) {
		s.mu.Lock()
		lastSeq := s.nextClientEventSeqLocked() - 1
		s.mu.Unlock()
		response, err := s.completeAIAgentOnboardingTask(ctx, principal, taskID, req, false)
		mutationErr = err
		s.mu.Lock()
		pending = nil
		for _, event := range s.events {
			if event.Seq > lastSeq {
				pending = append(pending, event)
			}
		}
		s.mu.Unlock()
		return response, err
	})
	if err != nil && mutationErr == nil && !errors.Is(err, ErrAIAgentMembershipChanged) {
		err = assignmentDependencyFailure(err)
	}
	if err == nil {
		s.mu.Lock()
		for _, event := range pending {
			s.fanoutClientEventLocked(event)
		}
		s.mu.Unlock()
	}
	return response, err
}
