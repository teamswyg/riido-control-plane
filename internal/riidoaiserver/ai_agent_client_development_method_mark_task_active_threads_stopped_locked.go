package riidoaiserver

import (
	"time"
)

func (s *DevelopmentAIAgentClientStore) markTaskActiveThreadsStoppedLocked(taskID string, kind AgentTaskCommentKind, message string) {
	now := time.Now().UTC()
	threads := s.taskThreads[taskID]
	for i := range threads {
		if !taskThreadHasActiveStream(threads[i]) {
			continue
		}
		threads[i].WorkStatus = AgentWorkStatusIdle
		threads[i].AssignmentState = AgentAssignmentStateStopped
		threads[i].CommentKind = kind
		threads[i].Message = message
		threads[i].CompletedAt = now
		if agent := s.agents[threads[i].AgentID]; agent.AgentID != "" && agent.AssignedTaskCount > 0 {
			agent.AssignedTaskCount--
			agent.Editability = editabilityForAssignedTasks(agent.AssignedTaskCount)
			if agent.AssignedTaskCount == 0 {
				agent.WorkStatus = AgentWorkStatusIdle
			}
			s.agents[agent.AgentID] = agent
		}
	}
	s.taskThreads[taskID] = threads
}
