package riidoaiserver

import (
	"context"
	"errors"
)

func (c *AIAgentPrivateTaskContextClient) getLegacyTaskDetailContext(ctx context.Context, req AIAgentTaskContextRequest, teamID, componentID, bearerToken string) (AIAgentTaskContext, error) {
	detail, err := c.getComponentDetail(ctx, teamID, componentID, bearerToken)
	if err != nil {
		return AIAgentTaskContext{}, err
	}
	if detail.ID != componentID {
		return AIAgentTaskContext{}, errors.New("riidoaiserver: task context component identity mismatch")
	}
	contextSnapshot := detail.toAIAgentTaskContext()
	contextSnapshot.TeamID = teamID
	if err := c.enrichTaskContextFor(ctx, teamID, componentID, bearerToken, &contextSnapshot, req); err != nil {
		return AIAgentTaskContext{}, err
	}
	return contextSnapshot, nil
}
