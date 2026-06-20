package riidoaiserver

import "strings"

func appendAIAgentTaskThreadMessagePrompt(prompt string, thread AIAgentTaskThreadRecord, req CreateAIAgentTaskThreadMessageRequest) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(prompt))
	b.WriteString("\n\n## Follow-up Thread Message\n")
	b.WriteString("- thread_id: ")
	b.WriteString(strings.TrimSpace(thread.ThreadID))
	b.WriteString("\n- previous_run_id: ")
	b.WriteString(strings.TrimSpace(thread.RunID))
	b.WriteString("\n- previous_work_status: ")
	b.WriteString(string(thread.WorkStatus))
	b.WriteString("\n- previous_assignment_state: ")
	b.WriteString(string(thread.AssignmentState))
	appendSourceMessageID(&b, req.SourceMessageID)
	appendPreviousThreadMessage(&b, thread.Message)
	b.WriteString("\n\n### New User Instruction\n")
	b.WriteString(strings.TrimSpace(req.Body))
	return strings.TrimSpace(b.String())
}

func appendSourceMessageID(b *strings.Builder, sourceMessageID string) {
	if sourceMessageID = strings.TrimSpace(sourceMessageID); sourceMessageID != "" {
		b.WriteString("\n- source_message_id: ")
		b.WriteString(sourceMessageID)
	}
}

func appendPreviousThreadMessage(b *strings.Builder, previousMessage string) {
	if previousMessage = strings.TrimSpace(previousMessage); previousMessage != "" {
		b.WriteString("\n\n### Previous Thread Message\n")
		b.WriteString(previousMessage)
	}
}
