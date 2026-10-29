package riidoaiserver

import (
	"errors"
	"strings"
)

func retainedTaskThreadMessages(source map[string][]AIAgentTaskThreadHistoryMessage) map[string][]AIAgentTaskThreadHistoryMessage {
	out := make(map[string][]AIAgentTaskThreadHistoryMessage, len(source))
	for threadID, messages := range source {
		out[threadID] = retainLatestThreadHistoryMessages(messages)
	}
	return out
}

func restoredTaskThreadMessages(source map[string][]AIAgentTaskThreadHistoryMessage) (map[string][]AIAgentTaskThreadHistoryMessage, error) {
	out := make(map[string][]AIAgentTaskThreadHistoryMessage, len(source))
	for threadID, messages := range source {
		threadID = strings.TrimSpace(threadID)
		if threadID == "" {
			return nil, errors.New("ai agent client snapshot task thread message thread_id is required")
		}
		out[threadID] = retainLatestThreadHistoryMessages(messages)
	}
	return out, nil
}
