package riidoaiserver

import (
	"context"
)

func (s *DevelopmentAIAgentClientStore) CreateAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest) (AgentClientRecordResponse, error) {
	return s.createAIAgent(ctx, principal, req, "")
}
