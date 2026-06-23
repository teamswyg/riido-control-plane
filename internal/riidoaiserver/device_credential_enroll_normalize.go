package riidoaiserver

import (
	"errors"
	"strings"
)

type normalizedDeviceEnrollment struct {
	principalID string
	workspaceID string
	displayName string
	machineID   string
	deviceID    string
}

func normalizeDeviceEnrollment(principal AuthorizationResult, workspaceID string, req EnrollDeviceRequest) (normalizedDeviceEnrollment, error) {
	out := normalizedDeviceEnrollment{
		principalID: strings.TrimSpace(principal.PrincipalID),
		workspaceID: strings.TrimSpace(workspaceID),
		displayName: strings.TrimSpace(req.DisplayName),
		machineID:   strings.TrimSpace(req.MachineID),
		deviceID:    strings.TrimSpace(req.DeviceID),
	}
	return completeDeviceEnrollment(out, principal)
}

func completeDeviceEnrollment(out normalizedDeviceEnrollment, principal AuthorizationResult) (normalizedDeviceEnrollment, error) {
	if out.workspaceID == "" {
		out.workspaceID = strings.TrimSpace(principal.WorkspaceID)
	}
	if out.principalID == "" {
		return normalizedDeviceEnrollment{}, errors.New("principal_id is required")
	}
	if out.workspaceID == "" {
		return normalizedDeviceEnrollment{}, errors.New("workspace_id is required")
	}
	if out.displayName == "" {
		out.displayName = "Riido Desktop"
	}
	return out, nil
}
