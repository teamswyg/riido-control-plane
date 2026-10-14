package riidoaiserver

func copyTaskThread(thread AIAgentTaskThreadRecord) AIAgentTaskThreadRecord {
	thread.AgentSnapshot = copyTaskThreadAgentSnapshot(thread.AgentSnapshot)
	thread.Lines = copyClientVisibleProgressLines(thread.Lines)
	thread.Message = clientVisibleTaskThreadMessage(thread)
	thread.ResultMessage = clientVisibleTaskThreadResultMessage(thread)
	thread.QueueDiagnostics = copyQueueDiagnostics(thread.QueueDiagnostics)
	thread.FailureDiagnostics = clientVisibleFailureDiagnostics(thread.FailureDiagnostics)
	return thread
}
