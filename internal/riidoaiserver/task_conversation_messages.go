package riidoaiserver

import (
	"sort"
	"strings"
)

func taskConversationMessages(thread AIAgentTaskThreadRecord, req CreateAIAgentTaskThreadMessageRequest, history []AIAgentTaskThreadHistoryRecord) []AIAgentTaskThreadHistoryMessage {
	var messages []AIAgentTaskThreadHistoryMessage
	seen := make(map[string]bool)
	linked := taskConversationLinkedThreadIDs(thread, history)
	for _, record := range history {
		if record.TaskID != thread.TaskID || (record.AgentID != thread.AgentID && !strings.HasPrefix(thread.ConversationID, "comment:") && !linked[record.ThreadID]) {
			continue
		}
		if thread.ConversationID != "" {
			if record.ConversationID != thread.ConversationID {
				continue
			}
		} else if record.ThreadID != thread.ThreadID || record.AgentID != thread.AgentID {
			continue
		}
		for _, message := range record.Messages {
			if message.Role != AIAgentTaskThreadMessageRoleUser && message.Role != AIAgentTaskThreadMessageRoleAgent {
				continue
			}
			if taskConversationMessageText(message, 2000) == "" {
				continue
			}
			if req.SourceMessageID != "" && message.SourceMessageID == req.SourceMessageID {
				continue
			}
			if message.MessageID != "" && seen[message.MessageID] {
				continue
			}
			seen[message.MessageID] = true
			messages = append(messages, message)
		}
	}
	sort.SliceStable(messages, func(i, j int) bool {
		if !messages[i].ObservedAt.Equal(messages[j].ObservedAt) {
			return messages[i].ObservedAt.Before(messages[j].ObservedAt)
		}
		return messages[i].Seq < messages[j].Seq
	})
	return messages
}
