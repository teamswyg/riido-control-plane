package riidoaiserver

import (
	"context"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) FindAIAgentTaskThreadByRunID(ctx context.Context, workspaceID, runID string) (AIAgentTaskThreadRecord, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadRecord{}, err
	}
	if strings.TrimSpace(runID) == "" || strings.TrimSpace(workspaceID) == "" {
		return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, threads := range s.taskThreads {
		for _, thread := range threads {
			if thread.RunID == runID && s.taskThreadWorkspaceIDLocked(thread) == workspaceID && thread.AssignmentID != "" {
				return copyTaskThread(thread), nil
			}
		}
	}
	return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
}

func (s *PersistentAIAgentClientStore) FindAIAgentTaskThreadByRunID(ctx context.Context, workspaceID, runID string) (AIAgentTaskThreadRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return AIAgentTaskThreadRecord{}, err
	}
	return s.DevelopmentAIAgentClientStore.FindAIAgentTaskThreadByRunID(ctx, workspaceID, runID)
}
