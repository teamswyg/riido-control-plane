package riidoaiserver

func aiAgentClientTaskThreadReactionRoute(taskBase string, n int) string {
	switch n {
	case 6:
		return taskBase + "/threads/{thread_id}/messages/{message_id}/reactions"
	case 7:
		return taskBase + "/threads/{thread_id}/messages/{message_id}/reactions/{emoji}"
	default:
		return ""
	}
}
