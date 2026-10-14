package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) agentForTaskStopLocked(principal AuthorizationResult, taskID, agentID, assignmentID string) (AgentClientRecord, bool) {
	if strings.TrimSpace(agentID) != "" {
		if agent, ok := s.visibleAgent(principal, agentID); ok {
			return agent, true
		}
		if thread, ok := s.taskThreadForStopTargetLocked(taskID, agentID, assignmentID); ok {
			return s.agentFromTaskThreadLocked(principal, thread)
		}
		return AgentClientRecord{}, false
	}
	if thread, ok := s.taskThreadForAssignmentAnyAgentLocked(taskID, assignmentID); ok {
		if agent, ok := s.visibleAgent(principal, thread.AgentID); ok {
			return agent, true
		}
		return s.agentFromTaskThreadLocked(principal, thread)
	}
	thread, ok := s.activeTaskThreadLocked(taskID)
	if !ok {
		return AgentClientRecord{}, false
	}
	if agent, ok := s.visibleAgent(principal, thread.AgentID); ok {
		return agent, true
	}
	return s.agentFromTaskThreadLocked(principal, thread)
}
