package riidoaiserver

import (
	"context"
	"strings"
	"time"
)

func taskRequestReceiptKey(principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) string {
	if strings.TrimSpace(req.RequestID) == "" {
		return ""
	}
	return identifiedAssignmentID(strings.TrimSpace(taskID), AssignRequest{WorkspaceID: principal.WorkspaceID, AgentID: strings.TrimSpace(req.AgentID), CreatedBy: principal.PrincipalID, RequestID: strings.TrimSpace(req.RequestID)})
}

func cloneTaskRequestReceipts(input map[string]AIAgentTaskActionResponse) map[string]AIAgentTaskActionResponse {
	result := make(map[string]AIAgentTaskActionResponse, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func (s *DevelopmentAIAgentClientStore) FindTaskRequestReceipt(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, bool, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskActionResponse{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.visibleAgent(principal, req.AgentID); !ok {
		return AIAgentTaskActionResponse{}, false, ErrAIAgentNotFound
	}
	response, found := s.taskRequestReceipts[taskRequestReceiptKey(principal, taskID, req)]
	if found {
		for _, thread := range s.taskThreads[taskID] {
			if thread.AssignmentID == response.AssignmentID {
				return actionResponseFromThread(thread, principal.WorkspaceID), true, nil
			}
		}
	}
	return response, found, nil
}

func (s *PersistentAIAgentClientStore) FindTaskRequestReceipt(ctx context.Context, principal AuthorizationResult, taskID string, req AssignAIAgentTaskRequest) (AIAgentTaskActionResponse, bool, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	s.snapshotReloadMu.Lock()
	s.snapshotLastReloadAt = time.Time{}
	s.snapshotReloadMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskActionResponse{}, false, err
	}
	return s.DevelopmentAIAgentClientStore.FindTaskRequestReceipt(ctx, principal, taskID, req)
}

func (s *DevelopmentAIAgentClientStore) rememberTaskRequestReceiptLocked(key string, response AIAgentTaskActionResponse) AIAgentTaskActionResponse {
	if key != "" {
		if s.taskRequestReceipts == nil {
			s.taskRequestReceipts = map[string]AIAgentTaskActionResponse{}
		}
		s.taskRequestReceipts[key] = response
	}
	return response
}
