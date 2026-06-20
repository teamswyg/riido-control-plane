package riidoaiserver

import "time"

func (s *DevelopmentAIAgentClientStore) assignmentProgressLineLocked(input assignmentEventInput, thread AIAgentTaskThreadRecord) AgentThreadProgressLine {
	messageCode, messageKey, messageArgs := progressLineMetadata(input.Metadata)
	message := input.Message
	if rendered, key, ok := renderProgressMessage(messageCode, messageArgs); ok {
		message = rendered
		if messageKey == "" {
			messageKey = key
		}
	}
	line := AgentThreadProgressLine{
		Seq:         s.nextThreadProgressSeqLocked(thread.TaskID, thread.ThreadID, input.Metadata),
		Message:     message,
		MessageCode: messageCode,
		MessageKey:  messageKey,
		MessageArgs: messageArgs,
		ObservedAt:  input.At,
	}
	if line.ObservedAt.IsZero() {
		line.ObservedAt = time.Now().UTC()
	}
	return line
}
