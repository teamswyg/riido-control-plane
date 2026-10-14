package riidoaiserver

func aiAgentClientWorkspaceAssignmentRoute(base string, segments []string) string {
	if len(segments) == 3 && segments[2] == "stop" {
		return base + "/agent-assignments/{agent_id}/stop"
	}
	return ""
}

func aiAgentClientThreadHTTPRoute(base string, segments []string) string {
	if len(segments) == 3 && segments[2] == "messages" {
		return base + "/threads/{thread_id}/messages"
	}
	return ""
}
