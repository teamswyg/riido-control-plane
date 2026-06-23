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
	for _, state := range []AgentAssignmentState{AgentAssignmentStateRunning, AgentAssignmentStateQueued} {
		if thread, ok := s.latestTaskThreadForAgentByStateLocked(taskID, agentID, state); ok {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) latestTaskThreadForAgentByStateLocked(taskID, agentID string, state AgentAssignmentState) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for i := len(threads) - 1; i >= 0; i-- {
		thread := threads[i]
		if thread.AgentID == agentID && thread.AssignmentState == state {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
