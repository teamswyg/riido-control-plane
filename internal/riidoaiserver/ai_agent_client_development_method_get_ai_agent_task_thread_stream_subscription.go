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
	threads := s.visibleTaskThreadsLocked(principal, taskID)
	filters := make([]AIAgentTaskThreadStreamTarget, 0, len(threads))
	for _, thread := range threads {
		if !taskThreadHasActiveStream(thread) {
			continue
		}
		filters = append(filters, AIAgentTaskThreadStreamTarget{
			AgentID:  thread.AgentID,
			ThreadID: thread.ThreadID,
			RunID:    thread.RunID,
		})
	}
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
