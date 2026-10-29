package riidoaiserver

import "strings"

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

// taskThreadStatusReactionTarget is the user message that started the run: the
// message with the thread source id, or else the first user message of the
// current run. It is "" when no stored thread message belongs to the run.
func taskThreadStatusReactionTarget(thread AIAgentTaskThreadRecord, messages []AIAgentTaskThreadHistoryMessage) string {
	source := strings.TrimSpace(thread.SourceMessageID)
	first := ""
	for _, message := range messages {
		if message.Role != AIAgentTaskThreadMessageRoleUser {
			continue
		}
		if source != "" && message.SourceMessageID == source {
			return message.MessageID
		}
		sameRun := message.AssignmentID == thread.AssignmentID && (message.RunID == "" || thread.RunID == "" || message.RunID == thread.RunID)
		if first == "" && sameRun {
			first = message.MessageID
		}
	}
	return first
}
