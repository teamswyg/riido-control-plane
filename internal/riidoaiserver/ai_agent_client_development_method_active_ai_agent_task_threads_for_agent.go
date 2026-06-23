package riidoaiserver

import (
	"context"
	"errors"
	"sort"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) ActiveAIAgentTaskThreadsForAgent(ctx context.Context, principal AuthorizationResult, agentID string) ([]AIAgentTaskThreadRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return nil, errors.New("agent_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	threads := s.activeTaskThreadsForAgentLocked(principal, agentID)
	sort.SliceStable(threads, func(i, j int) bool {
		return threads[i].StartedAt.After(threads[j].StartedAt)
	})
	return threads, nil
}

func (s *DevelopmentAIAgentClientStore) activeTaskThreadsForAgentLocked(principal AuthorizationResult, agentID string) []AIAgentTaskThreadRecord {
	out := []AIAgentTaskThreadRecord{}
	for taskID, source := range s.taskThreads {
		for i := range source {
			s.ensureTaskThreadAgentSnapshotLocked(&source[i], source[i].StartedAt)
			if source[i].AgentID != agentID || !taskThreadHasActiveStream(source[i]) || !s.taskThreadVisibleTo(principal, source[i]) {
				continue
			}
			out = append(out, copyTaskThread(source[i]))
		}
		s.taskThreads[taskID] = source
	}
	return out
}
