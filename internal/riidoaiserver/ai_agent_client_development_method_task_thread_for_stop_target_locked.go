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
	if thread, ok := s.stopTargetActiveTaskThreadForAgentLocked(taskID, agentID); ok {
		return thread, true
	}
	return s.latestTaskThreadForAgentLocked(taskID, agentID)
}

func (s *DevelopmentAIAgentClientStore) stopTargetActiveTaskThreadForAgentLocked(taskID, agentID string) (AIAgentTaskThreadRecord, bool) {
	return s.activeTaskThreadForAgentLocked(taskID, agentID)
}
