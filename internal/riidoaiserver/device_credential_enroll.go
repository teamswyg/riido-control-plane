package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"time"
)

func (s *DevelopmentAIAgentClientStore) EnrollDeviceCredential(ctx context.Context, principal AuthorizationResult, workspaceID string, req EnrollDeviceRequest) (EnrollDeviceResponse, error) {
	if err := ctx.Err(); err != nil {
		return EnrollDeviceResponse{}, err
	}
	enrollment, err := normalizeDeviceEnrollment(principal, workspaceID, req)
	if err != nil {
		return EnrollDeviceResponse{}, err
	}
	secret, err := newDeviceSecret()
	if err != nil {
		return EnrollDeviceResponse{}, err
	}
	hash := sha256.Sum256([]byte(secret))
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.deviceCredentials == nil {
		s.deviceCredentials = map[string]deviceCredentialRecord{}
	}
	deviceID, err := s.enrollmentDeviceIDLocked(enrollment)
	if err != nil {
		return EnrollDeviceResponse{}, err
	}
	s.deviceCredentials[deviceID] = deviceCredentialRecord{
		deviceID:         deviceID,
		machineID:        enrollment.machineID,
		secretHash:       hash,
		ownerPrincipalID: enrollment.principalID,
		workspaceID:      enrollment.workspaceID,
		displayName:      enrollment.displayName,
		issuedAt:         now,
	}
	s.upsertEnrolledDeviceLocked(DeviceRecord{
		DeviceID:              deviceID,
		OwnerPrincipalID:      enrollment.principalID,
		DisplayName:           enrollment.displayName,
		DaemonLastSeenAt:      now,
		ConnectedWorkspaceIDs: []string{enrollment.workspaceID},
	})
	return EnrollDeviceResponse{
		SchemaVersion:    DeviceCredentialSchemaVersion,
		DeviceID:         deviceID,
		DeviceSecret:     secret,
		OwnerPrincipalID: enrollment.principalID,
		WorkspaceID:      enrollment.workspaceID,
		DisplayName:      enrollment.displayName,
		IssuedAt:         now,
	}, nil
}
