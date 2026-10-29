package riidoaiserver

const (
	reactionEmojiStarted   = "1f440"
	reactionEmojiCompleted = "2705"
	reactionEmojiFailed    = "274c"
)

// agentStatusReactionEmoji is the one status reaction an agent keeps on the
// message or comment that started its run. Stopping, stopped, and unassigned
// runs keep none.
func agentStatusReactionEmoji(state AgentAssignmentState) string {
	switch state {
	case AgentAssignmentStateQueued, AgentAssignmentStateRunning, AgentAssignmentStateWaiting:
		return reactionEmojiStarted
	case AgentAssignmentStateCompleted:
		return reactionEmojiCompleted
	case AgentAssignmentStateFailed:
		return reactionEmojiFailed
	default:
		return ""
	}
}

// taskThreadReactionTarget keys the reactions of a history message. Agent
// messages are keyed by their run so reactions survive the message id change
// on each run state. Progress messages cannot be reacted to.
func taskThreadReactionTarget(message AIAgentTaskThreadHistoryMessage) (string, bool) {
	switch message.Role {
	case AIAgentTaskThreadMessageRoleUser:
		return message.MessageID, message.MessageID != ""
	case AIAgentTaskThreadMessageRoleAgent:
		return "run:" + message.AssignmentID + "/" + message.RunID, true
	default:
		return "", false
	}
}

// taskThreadStatusReactionTarget is the first user message of the current run,
// the message that started it. The thread source id is not used: a follow-up
// folded into a running run overwrites it. It is "" when no stored thread
// message belongs to the run (an assignment or a monolith comment).
func taskThreadStatusReactionTarget(thread AIAgentTaskThreadRecord, messages []AIAgentTaskThreadHistoryMessage) string {
	for _, message := range messages {
		sameRun := message.AssignmentID == thread.AssignmentID && (message.RunID == "" || thread.RunID == "" || message.RunID == thread.RunID)
		if message.Role == AIAgentTaskThreadMessageRoleUser && sameRun {
			return message.MessageID
		}
	}
	return ""
}

func findTaskThreadHistoryMessage(messages []AIAgentTaskThreadHistoryMessage, messageID string) (AIAgentTaskThreadHistoryMessage, bool) {
	for _, message := range messages {
		if messageID != "" && message.MessageID == messageID {
			return message, true
		}
	}
	return AIAgentTaskThreadHistoryMessage{}, false
}
