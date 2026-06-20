package riidoaiserver

import (
	"context"
	"errors"
	"slices"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ListAIAgentTaskThreads(ctx context.Context, principal AuthorizationResult, taskID string) (AIAgentTaskThreadCollectionResponse, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadCollectionResponse{}, err
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return AIAgentTaskThreadCollectionResponse{}, errors.New("task_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	threads := s.visibleTaskThreadsLocked(principal, taskID)
	response := AIAgentTaskThreadCollectionResponse{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Threads:       threads,
	}
	for _, thread := range slices.Backward(threads) {
		if !taskThreadHasActiveStream(thread) {
			continue
		}
		link := activeStreamLinkForThread(thread, strings.TrimSpace(principal.WorkspaceID))
		response.ActiveStream = &link
		break
	}
	return response, nil
}
