package riidoaiserver

import (
	"strings"
)

func (s *DevelopmentAIAgentClientStore) taskThreadByIDLocked(taskID, threadID string) (AIAgentTaskThreadRecord, bool) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return AIAgentTaskThreadRecord{}, false
	}
	for _, thread := range s.taskThreads[taskID] {
		if thread.ThreadID == threadID {
			return copyTaskThread(thread), true
		}
	}
	return AIAgentTaskThreadRecord{}, false
}
