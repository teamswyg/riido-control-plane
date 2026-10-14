package riidoaiserver

func hideQueuedThreadReplyMessage(response *AIAgentTaskActionResponse) {
	if response.CommentKind != AgentTaskCommentQueuedByBusyAgent {
		return
	}
	response.Message = ""
	response.ResultMessage = ""
}
