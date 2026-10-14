package riidoaiserver

import (
	"time"
)

func (s *DevelopmentAIAgentClientStore) markTaskAgentThreadsStopProjectionLocked(taskID, agentID string, response AIAgentTaskActionResponse, completed bool) {
	now := time.Now().UTC()
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].AgentID != agentID || !taskThreadHasActiveStream(threads[i]) {
			continue
		}
		applyTaskThreadStopProjection(&threads[i], response, completed, now)
	}
	s.taskThreads[taskID] = threads
}
