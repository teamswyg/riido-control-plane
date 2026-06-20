package riidoaiserver

import (
	"time"
)

func (s *DevelopmentAIAgentClientStore) markAgentTaskThreadsStoppedLocked(agentID string, kind AgentTaskCommentKind, message string) {
	now := time.Now().UTC()
	for taskID, threads := range s.taskThreads {
		for i := range threads {
			if threads[i].AgentID != agentID || !taskThreadHasActiveStream(threads[i]) {
				continue
			}
			threads[i].WorkStatus = AgentWorkStatusOffline
			threads[i].AssignmentState = AgentAssignmentStateStopped
			threads[i].CommentKind = kind
			threads[i].Message = message
			threads[i].ResultMessage = message
			threads[i].CompletedAt = now
		}
		s.taskThreads[taskID] = threads
	}
}
