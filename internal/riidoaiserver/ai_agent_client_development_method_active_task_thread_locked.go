package riidoaiserver

import (
	"slices"
)

func (s *DevelopmentAIAgentClientStore) activeTaskThreadLocked(taskID string) (AIAgentTaskThreadRecord, bool) {
	threads := s.taskThreads[taskID]
	for _, thread := range slices.Backward(threads) {
		if taskThreadHasActiveStream(thread) {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
