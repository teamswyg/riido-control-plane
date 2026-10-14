package riidoaiserver

import (
	"context"
)

func (s *PersistentAIAgentClientStore) observeTaskAgentMembershipRemoval(ctx context.Context, principal AuthorizationResult, taskID, agentID string) (AIAgentTaskThreadCollectionResponse, *taskAgentMembershipRemovalFence, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.forceReloadMembershipSnapshot(ctx); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, nil, err
	}
	threads, err := s.DevelopmentAIAgentClientStore.ListAIAgentTaskThreads(ctx, principal, taskID)
	if err != nil {
		return threads, nil, err
	}
	observation := observeTaskAgentMembershipRemoval(taskID, agentID, threads.Threads)
	s.mu.Lock()
	if agentID != "" {
		if _, visible := s.taskAgentForMembershipRemovalLocked(principal, taskID, agentID, ""); !visible {
			s.mu.Unlock()
			return threads, nil, ErrAIAgentNotFound
		}
	}
	observation.durableGeneration = s.taskAgentMemberships[taskID][agentID].DurableGeneration
	observation.intentGateAssignmentID = s.taskAgentMemberships[taskID][agentID].IntentGateAssignmentID
	s.mu.Unlock()
	return threads, observation, nil
}
