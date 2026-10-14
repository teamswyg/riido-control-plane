package riidoaiserver

func aiAgentClientTaskHTTPRoute(base string, segments []string) string {
	if len(segments) == 1 && segments[0] == "assigned-agent-profiles" {
		return base + "/tasks/assigned-agent-profiles"
	}
	if len(segments) < 2 {
		return ""
	}
	taskBase := base + "/tasks/{task_id}"
	switch segments[1] {
	case "assignable-agents", "assignment", "tool-approvals":
		return aiAgentClientTaskSimpleRoute(taskBase, segments)
	case "threads":
		return aiAgentClientTaskThreadsRoute(taskBase, segments)
	case "agent-assignments":
		return aiAgentClientTaskAssignmentRoute(taskBase, segments)
	case "thread-stream-subscription", "comments", "stop":
		return aiAgentClientTaskLeafRoute(taskBase, segments)
	default:
		return aiAgentClientTaskThreadMessageRoute(taskBase, segments)
	}
}

func aiAgentClientTaskSimpleRoute(taskBase string, segments []string) string {
	if len(segments) == 2 {
		return taskBase + "/" + segments[1]
	}
	if len(segments) == 4 && segments[1] == "tool-approvals" && segments[3] == "decision" {
		return taskBase + "/tool-approvals/{approval_id}/decision"
	}
	return ""
}

func aiAgentClientTaskAssignmentRoute(taskBase string, segments []string) string {
	if len(segments) == 2 {
		return taskBase + "/agent-assignments"
	}
	if len(segments) == 3 {
		return taskBase + "/agent-assignments/{agent_id}"
	}
	if len(segments) == 4 && segments[3] == "stop" {
		return taskBase + "/agent-assignments/{agent_id}/stop"
	}
	return ""
}

func aiAgentClientTaskLeafRoute(taskBase string, segments []string) string {
	if len(segments) == 2 {
		return taskBase + "/" + segments[1]
	}
	return ""
}

func aiAgentClientTaskThreadsRoute(taskBase string, segments []string) string {
	if len(segments) == 2 && segments[1] == "threads" {
		return taskBase + "/threads"
	}
	if len(segments) == 4 && segments[1] == "threads" && segments[3] == "messages" {
		return taskBase + "/threads/{thread_id}/messages"
	}
	return ""
}

func aiAgentClientTaskThreadMessageRoute(taskBase string, segments []string) string {
	return aiAgentClientTaskThreadsRoute(taskBase, segments)
}
