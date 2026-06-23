package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

func (s *DevelopmentAIAgentClientStore) createAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest, tmpColor string) (AgentClientRecordResponse, error) {
	if err := ctx.Err(); err != nil {
		return AgentClientRecordResponse{}, err
	}
	input, err := developmentCreateAgentInputFromRequest(req, tmpColor)
	if err != nil {
		return AgentClientRecordResponse{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	workspaceID := s.workspaceScope(principal)
	runtimeKind, runtimeModel, ok := runtimeSelectionFromDevices(s.visibleDevicesLocked(principal), input.RuntimeID, req.ModelID)
	if !ok {
		return AgentClientRecordResponse{}, errors.New("runtime_id or model_id is not available")
	}
	now := time.Now().UTC()
	agentID := uniqueAIAgentIDLocked(s.agents, "agent-"+principal.PrincipalID+"-"+input.RuntimeID)
	agent := AgentClientRecord{
		AgentID:             agentID,
		OwnerPrincipalID:    principal.PrincipalID,
		WorkspaceID:         workspaceID,
		IsOwnedByViewer:     true,
		Name:                input.Name,
		ProfileThumbnailURL: input.ProfileThumbnailURL,
		TmpColor:            input.TmpColor,
		Description:         input.Description,
		Instruction:         input.Instruction,
		Visibility:          input.Visibility,
		RuntimeID:           input.RuntimeID,
		RuntimeKind:         runtimeKind,
		ModelID:             runtimeModel.ModelID,
		ModelLabel:          runtimeModel.Label,
		WorkStatus:          AgentWorkStatusIdle,
		Editability:         AgentEditabilityEditable,
		AssignedTaskCount:   0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	s.agents[agent.AgentID] = agent
	return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: s.agentForPrincipal(agent, principal)}, nil
}
