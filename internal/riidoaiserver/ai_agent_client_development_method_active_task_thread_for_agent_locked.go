package riidoaiserver

import (
	"slices"
)

func (s *DevelopmentAIAgentClientStore) activeTaskThreadForAgentLocked(taskID, agentID string) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for _, thread := range slices.Backward(threads) {
		if thread.AgentID == agentID && taskThreadHasActiveStream(thread) {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
