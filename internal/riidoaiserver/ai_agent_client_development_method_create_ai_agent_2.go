package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

// createAIAgent preserves the creation preview color when provided. Older clients
// keep the least-used color selection by omitting profile_color.
func (s *DevelopmentAIAgentClientStore) createAIAgent(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest, tmpColor string) (AgentClientRecordResponse, error) {
	if req.ProfileColor != "" && !isAgentProfileColor(req.ProfileColor) {
		return AgentClientRecordResponse{}, errors.New("profile_color must be blue, orange, green, or purple")
	}
	return s.createAIAgentWithProfileColor(ctx, principal, req, tmpColor, req.ProfileColor, false)
}

// createAIAgentWithProfileColor creates an agent with a fixed profile color. An empty
// profileColor selects one of the least-used colors inside the store lock.
func (s *DevelopmentAIAgentClientStore) createAIAgentWithProfileColor(ctx context.Context, principal AuthorizationResult, req CreateAgentConfigurationRequest, tmpColor string, profileColor AgentProfileColor, onboardingCreated bool) (AgentClientRecordResponse, error) {
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
	if !isAgentProfileColor(profileColor) {
		profileColor = s.chooseNewAgentProfileColorLocked(workspaceID)
	}
	runtimeKind, runtimeModel, ok := runtimeSelectionFromDevices(s.visibleDevicesLocked(principal), input.RuntimeID, req.ModelID)
	if !ok {
		return AgentClientRecordResponse{}, errors.New("runtime_id or model_id is not available")
	}
	referenceFolders, err := s.normalizeReferenceFoldersLocked(req.ReferenceFolders, input.RuntimeID)
	if err != nil {
		return AgentClientRecordResponse{}, err
	}
	now := time.Now().UTC()
	agentID := uniqueAIAgentIDLocked(s.agents, "agent-"+principal.PrincipalID+"-"+input.RuntimeID)
	agent := AgentClientRecord{
		ReferenceFolders:    referenceFolders,
		AgentID:             agentID,
		OwnerPrincipalID:    principal.PrincipalID,
		WorkspaceID:         workspaceID,
		IsOwnedByViewer:     true,
		Name:                input.Name,
		ProfileThumbnailURL: input.ProfileThumbnailURL,
		TmpColor:            input.TmpColor,
		ProfileColor:        profileColor,
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
	if onboardingCreated {
		s.markOnboardingCreatedAgentLocked(agent.AgentID)
	}
	return AgentClientRecordResponse{SchemaVersion: SchemaVersion, Agent: s.agentForPrincipal(agent, principal)}, nil
}
