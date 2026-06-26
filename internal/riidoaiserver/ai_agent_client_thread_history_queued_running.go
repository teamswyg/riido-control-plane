package riidoaiserver

func runningTaskThreadHistoryConversations(
	threads []AIAgentTaskThreadHistoryRecord,
) map[string]struct{} {
	running := map[string]struct{}{}
	for _, thread := range threads {
		if threadIsActivelyRunning(thread) {
			running[thread.ConversationID] = struct{}{}
		}
	}
	return running
}

func taskThreadHistoryConversationIsRunning(
	conversationID string,
	running map[string]struct{},
) bool {
	_, ok := running[conversationID]
	return ok
}
