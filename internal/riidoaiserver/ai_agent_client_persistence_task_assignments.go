package riidoaiserver

import "context"

func (s *PersistentAIAgentClientStore) ListAIAgentTaskAssignableAgents(ctx context.Context, principal AuthorizationResult, taskID string) (AgentClientListResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AgentClientListResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskAssignableAgents(ctx, principal, taskID)
}

func (s *PersistentAIAgentClientStore) ListWorkspaceAssignedAgentProfiles(ctx context.Context, principal AuthorizationResult) (AssignedAgentProfileMapResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AssignedAgentProfileMapResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListWorkspaceAssignedAgentProfiles(ctx, principal)
}

func (s *PersistentAIAgentClientStore) ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskThreads(ctx, principal, taskID)
}

func (s *PersistentAIAgentClientStore) ListAIAgentTaskThreadHistory(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadHistoryCollectionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadHistoryCollectionResponse{}, err
	}
	return s.DevelopmentAIAgentClientStore.ListAIAgentTaskThreadHistory(ctx, principal, taskID)
}

func (s *PersistentAIAgentClientStore) ReconcileAIAgentActiveThreadProjections(ctx context.Context, principal AuthorizationResult, taskID string, reader AssignmentProjectionReader) (bool, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return false, err
	}
	changed, err := s.DevelopmentAIAgentClientStore.ReconcileAIAgentActiveThreadProjections(ctx, principal, taskID, reader)
	if err != nil || !changed {
		return changed, err
	}
	return changed, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) AssignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.AssignAIAgentTask(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) CreateAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.CreateAIAgentTaskAgentAssignment(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) UnassignAIAgentTask(ctx context.Context, principal AuthorizationResult, taskID string, req UnassignAIAgentTaskRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.UnassignAIAgentTask(ctx, principal, taskID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}

func (s *PersistentAIAgentClientStore) DeleteAIAgentTaskAgentAssignment(ctx context.Context, principal AuthorizationResult, taskID, agentID string, req AgentAssignmentActionRequest) (AIAgentTaskActionResponse, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, err
	}
	response, err := s.DevelopmentAIAgentClientStore.DeleteAIAgentTaskAgentAssignment(ctx, principal, taskID, agentID, req)
	if err != nil {
		return response, err
	}
	return response, s.saveSnapshot(ctx)
}
