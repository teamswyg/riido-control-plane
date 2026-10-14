package riidoaiserver

import (
	"slices"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) taskThreadForAssignmentLocked(taskID, agentID, assignmentID string) (AIAgentTaskThreadRecord, bool) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AIAgentTaskThreadRecord{}, false
	}
	threads := s.taskThreads[taskID]
	for _, thread := range slices.Backward(threads) {
		if thread.AgentID == agentID && thread.AssignmentID == assignmentID {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
