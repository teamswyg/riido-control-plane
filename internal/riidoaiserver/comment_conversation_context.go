package riidoaiserver

import "time"

type verifiedCommentContextMessage struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	AuthorID  string    `json:"author_principal_id"`
	CreatedAt time.Time `json:"created_at"`
}

func appendVerifiedCommentContext(history []AIAgentTaskThreadHistoryRecord, evidence verifiedCommentDelivery, selected AIAgentTaskThreadRecord) []AIAgentTaskThreadHistoryRecord {
	row := AIAgentTaskThreadHistoryRecord{TaskID: evidence.TaskID, AgentID: evidence.AgentID, ConversationID: selected.ConversationID}
	for _, m := range evidence.ConversationMessages {
		if m.ID == "" {
			continue
		}
		row.Messages = append(row.Messages, AIAgentTaskThreadHistoryMessage{
			MessageID:       taskThreadUserMessageID("", "", m.ID, m.Body, m.CreatedAt),
			SourceMessageID: m.ID, Role: AIAgentTaskThreadMessageRoleUser,
			AuthorPrincipalID: m.AuthorID, Body: m.Body, ObservedAt: m.CreatedAt,
		})
	}
	// Prefer the current stored comment body over its earlier invocation snapshot.
	return append([]AIAgentTaskThreadHistoryRecord{row}, history...)
}

func conversationMessageAuthor(message AIAgentTaskThreadHistoryMessage, history []AIAgentTaskThreadHistoryRecord) string {
	if message.Role != AIAgentTaskThreadMessageRoleAgent {
		return string(message.Role)
	}
	for _, row := range history {
		for _, candidate := range row.Messages {
			if message.MessageID != "" && candidate.MessageID == message.MessageID {
				return "agent (" + row.AgentID + ")"
			}
		}
	}
	return string(message.Role)
}
