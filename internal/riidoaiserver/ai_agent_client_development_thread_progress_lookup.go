package riidoaiserver

import (
	"slices"
	"strings"
)

func (s *DevelopmentAIAgentClientStore) rawTaskThreadForAssignmentLocked(
	taskID, agentID, assignmentID string,
) (AIAgentTaskThreadRecord, bool) {
	assignmentID = strings.TrimSpace(assignmentID)
	if assignmentID == "" {
		return AIAgentTaskThreadRecord{}, false
	}
	for _, thread := range slices.Backward(s.taskThreads[taskID]) {
		if thread.AgentID == agentID && thread.AssignmentID == assignmentID {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) rawActiveTaskThreadForAgentLocked(
	taskID, agentID string,
) (AIAgentTaskThreadRecord, bool) {
	for _, thread := range slices.Backward(s.taskThreads[taskID]) {
		if thread.AgentID == agentID && taskThreadHasActiveStream(thread) {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}

func (s *DevelopmentAIAgentClientStore) rawTaskThreadByIDLocked(
	taskID, threadID string,
) (AIAgentTaskThreadRecord, bool) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return AIAgentTaskThreadRecord{}, false
	}
	for _, thread := range s.taskThreads[taskID] {
		if thread.ThreadID == threadID {
			return thread, true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
