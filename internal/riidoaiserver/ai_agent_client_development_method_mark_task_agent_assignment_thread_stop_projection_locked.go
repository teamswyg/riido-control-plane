package riidoaiserver

import (
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) markTaskAgentAssignmentThreadStopProjectionLocked(taskID, agentID, assignmentID string, response AIAgentTaskActionResponse, completed bool) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return
	}
	now := time.Now().UTC()
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].AgentID != agentID || threads[i].AssignmentID != assignmentID || !taskThreadHasActiveStream(threads[i]) {
			continue
		}
		applyTaskThreadStopProjection(&threads[i], response, completed, now)
	}
	s.taskThreads[taskID] = threads
}
