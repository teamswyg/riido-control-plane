package riidoaiserver

import (
	"slices"
)

func (s *DevelopmentAIAgentClientStore) latestTaskThreadForAgentLocked(taskID, agentID string) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for _, thread := range slices.Backward(threads) {
		if thread.AgentID == agentID {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
