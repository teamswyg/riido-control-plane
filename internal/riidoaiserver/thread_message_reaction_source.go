package riidoaiserver

// A follow-up without request_id or source_message_id still has a generated
// user message. A new execution without an explicit source reacts to it.
func (s *DevelopmentAIAgentClientStore) bindGeneratedReactionSourceLocked(taskID, threadID, parentThreadID, messageID string) {
	if parentThreadID == "" || messageID == "" {
		return
	}
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].ThreadID != threadID || threads[i].ReactionSourceMessageID == nil || *threads[i].ReactionSourceMessageID != "" {
			continue
		}
		source := messageID
		threads[i].ReactionSourceMessageID = &source
	}
}

// The projected agent message replaces a stored copy with the same id and
// keeps the member reactions stored on it.
func upsertProjectionMessage(out []AIAgentTaskThreadHistoryMessage, message AIAgentTaskThreadHistoryMessage) []AIAgentTaskThreadHistoryMessage {
	for i := range out {
		if out[i].MessageID == message.MessageID {
			message.Reactions = out[i].Reactions
			out[i] = message
			return out
		}
	}
	return append(out, message)
}
