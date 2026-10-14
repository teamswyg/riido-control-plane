package riidoaiserver

import (
	"strconv"
)

func uniqueAIAgentIDLocked(agents map[string]AgentClientRecord, seed string) string {
	base := slugAIAgentIDComponent(seed)
	if base == "" {
		base = "agent"
	}
	if _, exists := agents[base]; !exists {
		return base
	}
	for i := 2; ; i++ {
		candidate := base + "-" + strconv.Itoa(i)
		if _, exists := agents[candidate]; !exists {
			return candidate
		}
	}
}
