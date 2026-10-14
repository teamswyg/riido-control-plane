package riidoaiserver

import (
	"errors"
	"strings"
)

func applyAgentConfigurationPatch(agent *AgentClientRecord, req UpdateAgentConfigurationRequest) error {
	if strings.TrimSpace(req.Name) != "" {
		agent.Name = strings.TrimSpace(req.Name)
	}
	if err := applyAgentOptionalTextPatch(agent, req); err != nil {
		return err
	}
	if req.Visibility == "" {
		return nil
	}
	if req.Visibility != AgentVisibilityPublic && req.Visibility != AgentVisibilityPrivate {
		return errors.New("visibility must be public or private")
	}
	agent.Visibility = req.Visibility
	return nil
}

func applyAgentOptionalTextPatch(agent *AgentClientRecord, req UpdateAgentConfigurationRequest) error {
	if req.ProfileThumbnailURL != nil {
		thumbnailURL, err := normalizeAgentProfileThumbnailURL(*req.ProfileThumbnailURL)
		if err != nil {
			return err
		}
		agent.ProfileThumbnailURL = thumbnailURL
	}
	if req.Description != nil {
		if err := validateAgentDescription(*req.Description); err != nil {
			return err
		}
		agent.Description = *req.Description
	}
	if req.Instruction != nil {
		if err := validateAgentInstruction(*req.Instruction); err != nil {
			return err
		}
		agent.Instruction = *req.Instruction
	}
	return nil
}
