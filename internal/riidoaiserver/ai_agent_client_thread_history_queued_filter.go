package riidoaiserver

func historyThreadMessageShouldDropQueued(
	message AIAgentTaskThreadHistoryMessage,
) bool {
	return historyMessageIsQueuedStatus(message)
}
