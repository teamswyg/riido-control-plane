package riidoaiserver

import "time"

func developmentSeedAgents(now time.Time) map[string]AgentClientRecord {
	return map[string]AgentClientRecord{
		"agent-owned-codex":     developmentOwnedCodexAgent(now),
		"agent-owned-claude":    developmentOwnedClaudeAgent(now),
		"agent-public-openclaw": developmentPublicOpenClawAgent(now),
		"agent-private-cursor":  developmentPrivateCursorAgent(now),
	}
}
