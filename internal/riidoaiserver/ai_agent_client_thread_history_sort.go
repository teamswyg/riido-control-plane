package riidoaiserver

import "sort"

func sortTaskThreadHistoryMessages(messages []AIAgentTaskThreadHistoryMessage) {
	if len(messages) < 2 || taskThreadHistoryMessagesSorted(messages) {
		return
	}
	sort.SliceStable(messages, func(i, j int) bool {
		return taskThreadHistoryMessageLess(messages[i], messages[j])
	})
}

func taskThreadHistoryMessagesSorted(messages []AIAgentTaskThreadHistoryMessage) bool {
	for i := 1; i < len(messages); i++ {
		if taskThreadHistoryMessageLess(messages[i], messages[i-1]) {
			return false
		}
	}
	return true
}

func taskThreadHistoryMessageLess(left, right AIAgentTaskThreadHistoryMessage) bool {
	if !left.ObservedAt.Equal(right.ObservedAt) {
		return left.ObservedAt.Before(right.ObservedAt)
	}
	if messageRoleRank(left.Role) != messageRoleRank(right.Role) {
		return messageRoleRank(left.Role) < messageRoleRank(right.Role)
	}
	if left.Seq != right.Seq {
		return left.Seq < right.Seq
	}
	return left.MessageID < right.MessageID
}

func messageRoleRank(role AIAgentTaskThreadMessageRole) int {
	switch role {
	case AIAgentTaskThreadMessageRoleUser:
		return 0
	case AIAgentTaskThreadMessageRoleProgress:
		return 1
	case AIAgentTaskThreadMessageRoleAgent:
		return 2
	default:
		return 3
	}
}
