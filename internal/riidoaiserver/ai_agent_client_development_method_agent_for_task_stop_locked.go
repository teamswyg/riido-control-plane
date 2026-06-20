package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) agentForTaskStopLocked(principal AuthorizationResult, taskID, agentID, assignmentID string) (AgentClientRecord, bool) {
	if strings.TrimSpace(agentID) != "" {
		return s.visibleAgent(principal, agentID)
	}
	if thread, ok := s.taskThreadForAssignmentAnyAgentLocked(taskID, assignmentID); ok {
		return s.visibleAgent(principal, thread.AgentID)
	}
	thread, ok := s.activeTaskThreadLocked(taskID)
	if !ok {
		return AgentClientRecord{}, false
	}
	return s.visibleAgent(principal, thread.AgentID)
}
