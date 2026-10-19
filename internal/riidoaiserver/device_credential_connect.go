package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"
)

// CredentialDeviceConnector links the current authorized user to the device
// actually running on this machine, without rotating its original credential.
type CredentialDeviceConnector interface {
	ConnectAIAgentDeviceWithCredential(context.Context, AuthorizationResult, string, string) (DeviceRecord, error)
}

func (s *DevelopmentAIAgentClientStore) ConnectAIAgentDeviceWithCredential(ctx context.Context, principal AuthorizationResult, deviceID, secret string) (DeviceRecord, error) {
	if err := ctx.Err(); err != nil {
		return DeviceRecord{}, err
	}
	deviceID = strings.TrimSpace(deviceID)
	secret = strings.TrimSpace(secret)
	if deviceID == "" || secret == "" {
		return DeviceRecord{}, ErrAuthorizationUnauthenticated
	}
	if strings.TrimSpace(principal.PrincipalID) == "" || strings.TrimSpace(principal.WorkspaceID) == "" {
		return DeviceRecord{}, ErrAuthorizationForbidden
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	credential, ok := s.deviceCredentials[deviceID]
	got := sha256.Sum256([]byte(secret))
	if !ok || subtle.ConstantTimeCompare(got[:], credential.secretHash[:]) != 1 {
		return DeviceRecord{}, ErrAuthorizationUnauthenticated
	}
	for i := range s.devices {
		if s.devices[i].DeviceID != deviceID {
			continue
		}
		s.devices[i].ConnectedWorkspaceIDs = addConnectedWorkspace(s.devices[i].ConnectedWorkspaceIDs, principal.WorkspaceID)
		s.upsertDeviceConnectionGrantLocked(deviceID, principal, time.Now().UTC())
		return copyDevice(s.devices[i]), nil
	}
	return DeviceRecord{}, ErrAuthorizationForbidden
}

func (s *PersistentAIAgentClientStore) ConnectAIAgentDeviceWithCredential(ctx context.Context, principal AuthorizationResult, deviceID, secret string) (DeviceRecord, error) {
	s.operationMu.Lock()
	defer s.operationMu.Unlock()
	if err := s.reloadSnapshot(ctx); err != nil {
		return DeviceRecord{}, err
	}
	before := s.desktopPresenceGrantCheckpoint()
	device, err := s.DevelopmentAIAgentClientStore.ConnectAIAgentDeviceWithCredential(ctx, principal, deviceID, secret)
	if err != nil {
		return DeviceRecord{}, err
	}
	return device, s.commitDesktopPresenceGrants(ctx, before)
}
