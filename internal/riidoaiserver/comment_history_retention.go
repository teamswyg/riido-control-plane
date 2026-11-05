package riidoaiserver

import "strings"

// Conversation messages are durable user-visible records, not a progress tail.
func durableConversationHistory(thread AIAgentTaskThreadRecord, messages []AIAgentTaskThreadHistoryMessage) bool {
	if strings.HasPrefix(thread.ConversationID, "comment:") {
		return true
	}
	for _, message := range messages {
		if strings.HasPrefix(message.MessageID, "native:") {
			return true
		}
	}
	return false
}

func retainedConversationMessages(threads map[string][]AIAgentTaskThreadRecord, messages map[string][]AIAgentTaskThreadHistoryMessage) map[string][]AIAgentTaskThreadHistoryMessage {
	durable := map[string]bool{}
	for _, rows := range threads {
		for _, thread := range rows {
			durable[thread.ThreadID] = durableConversationHistory(thread, messages[thread.ThreadID])
		}
	}
	out := make(map[string][]AIAgentTaskThreadHistoryMessage, len(messages))
	for id, rows := range messages {
		if durable[id] || durableConversationHistory(AIAgentTaskThreadRecord{}, rows) {
			out[id] = copyTaskThreadHistoryMessages(rows)
		} else {
			out[id] = retainLatestThreadHistoryMessages(rows)
		}
	}
	return out
}

func snapshotHasConversationHistory(snapshot AIAgentClientSnapshot) bool {
	for _, rows := range snapshot.TaskThreads {
		for _, thread := range rows {
			if durableConversationHistory(thread, snapshot.TaskThreadMessages[thread.ThreadID]) {
				return true
			}
		}
	}
	for _, rows := range snapshot.TaskThreadMessages {
		if durableConversationHistory(AIAgentTaskThreadRecord{}, rows) {
			return true
		}
	}
	return false
}
