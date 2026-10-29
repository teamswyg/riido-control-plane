package riidoaiserver

import "strings"

func appendTaskConversationContext(out *strings.Builder, thread AIAgentTaskThreadRecord, req CreateAIAgentTaskThreadMessageRequest, history []AIAgentTaskThreadHistoryRecord) bool {
	messages := taskConversationMessages(thread, req, history)
	if len(messages) == 0 {
		return false
	}
	start := max(0, len(messages)-8)
	var parts []string
	remaining := conversationContextLimit
	for i := len(messages) - 1; i >= start && remaining > 128; i-- {
		message := messages[i]
		text := taskConversationMessageText(message, min(2000, remaining-64))
		if text == "" {
			continue
		}
		author := string(message.Role)
		if thread.ConversationID != "" {
			author = conversationMessageAuthor(message, history)
		}
		part := "\n### Previous " + author + " message\n" + text + "\n"
		parts = append(parts, part)
		remaining -= len([]rune(part))
	}
	if len(parts) == 0 {
		return false
	}
	out.WriteString("\n\n## Recent conversation context\nReference history; the New User Instruction below is authoritative.\n")
	if len(parts) < len(messages) {
		out.WriteString("[Earlier conversation messages omitted by context budget.]\n")
	}
	for i := len(parts) - 1; i >= 0; i-- {
		out.WriteString(parts[i])
	}
	return true
}

// Keep the eligibility filter and final rendering identical so empty rich-text
// status rows cannot consume the finite recent-message budget.
func taskConversationMessageText(message AIAgentTaskThreadHistoryMessage, limit int) string {
	if message.Role == AIAgentTaskThreadMessageRoleAgent && message.ResultMessage != "" {
		return taskContextText(message.ResultMessage, "markdown", limit)
	}
	return taskMessageContextText(message.Body, limit)
}
