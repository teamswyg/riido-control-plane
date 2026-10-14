package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskThreadMessage(ctx context.Context, principal AuthorizationResult, taskID, threadID string, req CreateAIAgentTaskThreadMessageRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	// Replies that only append dialogue do not own membership.
	if req.membershipPublicationGeneration == "" && req.AssignmentID == "" {
		if err := s.reloadSnapshot(ctx); err != nil {
			return AIAgentTaskActionResponse{}, err
		}
		response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
		if err != nil {
			return response, err
		}
		return response, s.saveSnapshot(ctx)
	}
	return s.mutateTaskAgentMembership(ctx, taskID, "", req.membershipPublicationGeneration, nil, true, func() (AIAgentTaskActionResponse, error) {
		return s.DevelopmentAIAgentClientStore.CreateAIAgentTaskThreadMessage(ctx, principal, taskID, threadID, req)
	})
}
