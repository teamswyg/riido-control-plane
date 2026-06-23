package riidoaiserver

import "sort"

func sortTaskThreadHistoryMessages(messages []AIAgentTaskThreadHistoryMessage) {
	sort.SliceStable(messages, func(i, j int) bool {
		return messages[i].ObservedAt.Before(messages[j].ObservedAt)
	})
}
