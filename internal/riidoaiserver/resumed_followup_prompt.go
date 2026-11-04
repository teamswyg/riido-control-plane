package riidoaiserver

import (
	"strings"
	"time"
)

// Native resume already restores this Agent's conversation. Only messages that
// arrived outside that session after its previous request need to be supplied.
func resumedFollowupPrompt(body string, thread AIAgentTaskThreadRecord, req CreateAIAgentTaskThreadMessageRequest, history []AIAgentTaskThreadHistoryRecord, since time.Time) string {
	instruction := taskMessageContextText(body, taskDocumentContextLimit)
	var out strings.Builder
	if !appendTaskConversationContext(&out, thread, req, resumedFollowupHistory(thread, history, since)) {
		return instruction
	}
	out.WriteString("\n### New User Instruction\n" + assignmentInstructionHeading + "\n")
	out.WriteString(instruction)
	return strings.TrimSpace(out.String())
}

func resumedFollowupHistory(thread AIAgentTaskThreadRecord, history []AIAgentTaskThreadHistoryRecord, since time.Time) []AIAgentTaskThreadHistoryRecord {
	known := map[string]bool{}
	for _, row := range history {
		if row.ThreadID == thread.ThreadID && row.AgentID == thread.AgentID {
			for _, message := range row.Messages {
				if message.SourceMessageID != "" {
					known[message.SourceMessageID] = true
				}
			}
		}
	}
	var delta []AIAgentTaskThreadHistoryRecord
	for _, row := range history {
		record := row
		record.Messages = nil
		if row.ThreadID != "" && row.AgentID == thread.AgentID {
			delta = append(delta, record) // Keep authorized parent links, without native history.
			continue
		}
		for _, message := range row.Messages {
			// Unknown timestamps are retained; they cannot establish that a
			// message predates this session's last request.
			if known[message.SourceMessageID] || (!message.ObservedAt.IsZero() && message.ObservedAt.Before(since)) {
				continue
			}
			record.Messages = append(record.Messages, message)
		}
		delta = append(delta, record)
	}
	return delta
}
