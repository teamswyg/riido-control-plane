package riidoaiserver

import (
	"errors"
	"strings"
)

func developmentCreateAgentInputFromRequest(req CreateAgentConfigurationRequest, tmpColor string) (developmentCreateAgentInput, error) {
	input := developmentCreateAgentInput{
		Name:       strings.TrimSpace(req.Name),
		RuntimeID:  strings.TrimSpace(req.RuntimeID),
		TmpColor:   strings.TrimSpace(tmpColor),
		Visibility: req.Visibility,
	}
	if input.Name == "" {
		return developmentCreateAgentInput{}, errors.New("name is required")
	}
	if input.RuntimeID == "" {
		return developmentCreateAgentInput{}, errors.New("runtime_id is required")
	}
	if input.Visibility != AgentVisibilityPublic && input.Visibility != AgentVisibilityPrivate {
		return developmentCreateAgentInput{}, errors.New("visibility must be public or private")
	}
	if err := fillCreateAgentOptionalFields(&input, req); err != nil {
		return developmentCreateAgentInput{}, err
	}
	return input, nil
}

func fillCreateAgentOptionalFields(input *developmentCreateAgentInput, req CreateAgentConfigurationRequest) error {
	if req.ProfileThumbnailURL != nil {
		thumbnailURL, err := normalizeAgentProfileThumbnailURL(*req.ProfileThumbnailURL)
		if err != nil {
			return err
		}
		input.ProfileThumbnailURL = thumbnailURL
	}
	if req.Description != nil {
		if err := validateAgentDescription(*req.Description); err != nil {
			return err
		}
		input.Description = *req.Description
	}
	if req.Instruction != nil {
		if err := validateAgentInstruction(*req.Instruction); err != nil {
			return err
		}
		input.Instruction = *req.Instruction
	}
	return nil
}
