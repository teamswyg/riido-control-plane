package riidoaiserver

import "sort"

func sortTaskThreadHistoryMessages(messages []AIAgentTaskThreadHistoryMessage) {
	sort.SliceStable(messages, func(i, j int) bool {
		left := messages[i]
		right := messages[j]
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
	})
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
