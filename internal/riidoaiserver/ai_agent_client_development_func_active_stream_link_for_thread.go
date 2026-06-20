package riidoaiserver

func activeStreamLinkForThread(thread AIAgentTaskThreadRecord, workspaceID string) AIAgentTaskThreadStreamLink {
	return AIAgentTaskThreadStreamLink{
		Rel:       "agent_thread_progress_stream",
		Href:      aiAgentClientEventStreamHref(workspaceID),
		EventType: AgentClientEventThreadProgress,
		TaskID:    thread.TaskID,
		ThreadID:  thread.ThreadID,
		RunID:     thread.RunID,
	}
}
