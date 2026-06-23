package riidoaiserver

import (
	"context"
	"errors"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) GetAIAgentTaskThreadStreamSubscription(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadStreamSubscriptionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadStreamSubscriptionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return AIAgentTaskThreadStreamSubscriptionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	filters := s.activeTaskThreadStreamTargetsLocked(principal, taskID)
	return AIAgentTaskThreadStreamSubscriptionResponse{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Stream: AIAgentTaskEventStreamLink{
			Rel:       "agent_thread_progress_stream",
			Href:      aiAgentClientEventStreamHref(strings.TrimSpace(principal.WorkspaceID)),
			EventType: AgentClientEventThreadProgress,
		},
		ActiveThreadFilters: filters,
	}, nil
}
