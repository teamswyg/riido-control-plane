package riidoaiserver

func (s Server) runtimeBoundAssignableAgents(agents []AgentClientRecord) []AgentClientRecord {
	assignable := make([]AgentClientRecord, 0, len(agents))
	for _, agent := range agents {
		if _, _, err := s.resolveAgentRuntimeFact(agent.AgentID); err == nil {
			assignable = append(assignable, agent)
		}
	}
	return assignable
}
