package riidoaiserver

import "context"

type AIAgentActiveTaskThreadsForAgentResolver interface {
	ActiveAIAgentTaskThreadsForAgent(ctx context.Context, principal AuthorizationResult, agentID string) ([]AIAgentTaskThreadRecord, error)
}
