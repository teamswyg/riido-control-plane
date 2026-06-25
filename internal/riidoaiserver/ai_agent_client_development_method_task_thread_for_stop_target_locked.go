package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) taskThreadForStopTargetLocked(taskID, agentID, assignmentID string) (AIAgentTaskThreadRecord, bool) {
	if thread, ok := s.taskThreadForAssignmentLocked(taskID, agentID, assignmentID); ok {
		return thread, true
	}
	if strings.TrimSpace(assignmentID) != "" {
		return AIAgentTaskThreadRecord{}, false
	}
	return s.stopTargetActiveTaskThreadForAgentLocked(taskID, agentID)
}

func (s *DevelopmentAIAgentClientStore) taskThreadForUnassignTargetLocked(taskID, agentID, assignmentID string) (AIAgentTaskThreadRecord, bool) {
	if thread, ok := s.taskThreadForStopTargetLocked(taskID, agentID, assignmentID); ok {
		return thread, true
	}
	if strings.TrimSpace(assignmentID) != "" {
		return AIAgentTaskThreadRecord{}, false
	}
	return s.latestTaskThreadForAgentLocked(taskID, agentID)
}

func (s *DevelopmentAIAgentClientStore) stopTargetActiveTaskThreadForAgentLocked(taskID, agentID string) (AIAgentTaskThreadRecord, bool) {
	return s.activeTaskThreadForAgentLocked(taskID, agentID)
}
