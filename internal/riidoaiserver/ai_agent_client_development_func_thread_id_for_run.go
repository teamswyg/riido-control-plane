package riidoaiserver

func threadIDForRun(taskID, agentID, runID string) string {
	return "thread-" + slugAIAgentIDComponent(taskID+"-"+agentID+"-"+runID)
}
