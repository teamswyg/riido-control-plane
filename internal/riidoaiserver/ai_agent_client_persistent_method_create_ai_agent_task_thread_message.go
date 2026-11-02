package riidoaiserver

import (
	"context"
	"errors"
)

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	// Replies that only append dialog do not own membership.
	if req.membershipPublicationGeneration == "" && req.AssignmentID == "" {
		var response AIAgentTaskActionResponse
		var err error
		for attempt := 0; attempt < threadSnapshotSaveAttempts; attempt++ {
			if err := s.reloadSnapshot(ctx); err != nil {
				return AIAgentTaskActionResponse{}, err
			}
			response, err = s.DevelopmentAIAgentClientStore.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
			if err != nil {
				return response, err
			}
			err = s.saveSnapshot(ctx)
			if !errors.Is(err, errTaskAgentSnapshotChanged) && !errors.Is(err, errAIAgentSnapshotConflict) {
				return response, err
			}
		}
		return response, err
	}
	return s.mutateTaskAgentMembership(ctx, taskID, "", req.membershipPublicationGeneration, nil, true, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
	})
}
