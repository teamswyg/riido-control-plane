package riidoaiserver

func activeStreamLinkForThread(thread AIAgentTaskThreadRecord, workspaceID string) AIAgentTaskThreadStreamLink {
	return activeStreamLinkForThreadHref(thread, aiAgentClientEventStreamHref(workspaceID))
}

func activeStreamLinkForThreadHref(thread AIAgentTaskThreadRecord, href string) AIAgentTaskThreadStreamLink {
	return AIAgentTaskThreadStreamLink{
		Rel:       "agent_thread_progress_stream",
		Href:      href,
		EventType: AgentClientEventThreadProgress,
		TaskID:    thread.TaskID,
		ThreadID:  thread.ThreadID,
		RunID:     thread.RunID,
	}
}
