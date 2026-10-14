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
		body := message.Body
		format := "message"
		if message.Role == AIAgentTaskThreadMessageRoleAgent {
			if message.ResultMessage != "" {
				body = message.ResultMessage
				format = "markdown"
			}
		}
		text := taskContextText(body, format, min(2000, remaining-64))
		if format == "message" {
			text = taskMessageContextText(body, min(2000, remaining-64))
		}
		if text == "" {
			continue
		}
		part := "\n### Previous " + string(message.Role) + " message\n" + text + "\n"
		parts = append(parts, part)
		remaining -= len([]rune(part))
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
