package riidoaiserver

func selectAssignableAgent(agents []AgentClientRecord, agentID string) (AgentClientRecord, error) {
	for _, agent := range agents {
		if agent.AgentID == agentID {
			return agent, nil
		}
	}
	return AgentClientRecord{}, ErrAIAgentNotFound
}
