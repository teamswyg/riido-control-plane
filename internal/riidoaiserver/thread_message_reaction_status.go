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
// messages are keyed by their assignment so reactions survive the message id
// change on each run state; a progress batch without run_id rewrites the run
// id, so it is used only when no assignment exists. Progress messages cannot
// be reacted to.
func taskThreadReactionTarget(message AIAgentTaskThreadHistoryMessage) (string, bool) {
	switch message.Role {
	case AIAgentTaskThreadMessageRoleUser:
		return message.MessageID, message.MessageID != ""
	case AIAgentTaskThreadMessageRoleAgent:
		if message.AssignmentID != "" {
			return "run:" + message.AssignmentID, true
		}
		return "run:/" + message.RunID, message.RunID != ""
	default:
		return "", false
	}
}

// taskThreadStatusReactionTarget is the first user message of the current
// assignment, the message that started it. Neither the thread source id nor
// the run id is used: a follow-up folded into a running run overwrites the
// first and a progress batch without run_id rewrites the second. It is "" when
// no stored thread message belongs to the assignment (an assignment created
// outside the thread or a monolith comment).
func taskThreadStatusReactionTarget(thread AIAgentTaskThreadRecord, messages []AIAgentTaskThreadHistoryMessage) string {
	if thread.AssignmentID == "" {
		return ""
	}
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleUser && message.AssignmentID == thread.AssignmentID {
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
