package riidoaiserver

import (
	"context"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) FindAIAgentTaskThreadByID(ctx context.Context, workspaceID, threadID string) (AIAgentTaskThreadRecord, error) {
	if err := ctx.Err(); err != nil {
		return AIAgentTaskThreadRecord{}, err
	}
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	workspaceID = s.workspaceScope(AuthorizationResult{WorkspaceID: workspaceID})
	for taskID, source := range s.taskThreads {
		for i := range source {
			if source[i].ThreadID != threadID {
				continue
			}
			s.ensureTaskThreadAgentSnapshotLocked(&source[i], source[i].StartedAt)
			s.taskThreads[taskID] = source
			if s.taskThreadWorkspaceIDLocked(source[i]) != workspaceID {
				return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
			}
			return copyTaskThread(source[i]), nil
		}
	}
	return AIAgentTaskThreadRecord{}, ErrAIAgentNotFound
}
