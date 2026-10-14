package riidoaiserver

import "slices"

func aiAgentIDs(agents []AgentClientRecord) []string {
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		ids = append(ids, agent.AgentID)
	}
	return ids
}

func aiAgentIDsWithName(agents []AgentClientRecord, name string) []string {
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		if agent.Name == name {
			ids = append(ids, agent.AgentID)
		}
	}
	return ids
}

func containsString(values []string, want string) bool {
	return slices.Contains(values, want)
}
