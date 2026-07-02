package riidoaiserver

func historyRecordWithMessages(
	threadID string,
	conversationID string,
	messages []AIAgentTaskThreadHistoryMessage,
) AIAgentTaskThreadHistoryRecord {
	return AIAgentTaskThreadHistoryRecord{
		ThreadID: threadID, ConversationID: conversationID, Messages: messages,
	}
}

func historyHasAgentCommentKind(messages []AIAgentTaskThreadHistoryMessage, kind AgentTaskCommentKind) bool {
	for _, message := range messages {
		if message.Role == AIAgentTaskThreadMessageRoleAgent && message.CommentKind == kind {
			return true
		}
	}
	return false
}
