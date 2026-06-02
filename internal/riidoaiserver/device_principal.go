package riidoaiserver

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const DeviceCredentialSchemaVersion = "riido-device-credential.v1"

type EnrollDeviceRequest struct {
	DisplayName string `json:"display_name,omitempty"`
	Platform    string `json:"platform,omitempty"`
	AppVersion  string `json:"app_version,omitempty"`
}

type EnrollDeviceResponse struct {
	SchemaVersion    string    `json:"schema_version"`
	DeviceID         string    `json:"device_id"`
	DeviceSecret     string    `json:"device_secret"`
	OwnerPrincipalID string    `json:"owner_principal_id"`
	WorkspaceID      string    `json:"workspace_id,omitempty"`
	DisplayName      string    `json:"display_name,omitempty"`
	IssuedAt         time.Time `json:"issued_at"`
}

type DeviceCredentialStore interface {
	EnrollDeviceCredential(ctx context.Context, principal AuthorizationResult, workspaceID string, req EnrollDeviceRequest) (EnrollDeviceResponse, error)
	AuthorizeDeviceCredential(ctx context.Context, deviceID, deviceSecret string, req AuthorizationRequest) (AuthorizationResult, error)
}

type deviceCredentialRecord struct {
	deviceID         string
	secretHash       [sha256.Size]byte
	ownerPrincipalID string
	workspaceID      string
	displayName      string
	issuedAt         time.Time
}

func (s *DevelopmentAIAgentClientStore) EnrollDeviceCredential(ctx context.Context, principal AuthorizationResult, workspaceID string, req EnrollDeviceRequest) (EnrollDeviceResponse, error) {
	if err := ctx.Err(); err != nil {
		return EnrollDeviceResponse{}, err
	}
	principal.PrincipalID = strings.TrimSpace(principal.PrincipalID)
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		workspaceID = strings.TrimSpace(principal.WorkspaceID)
	}
	if principal.PrincipalID == "" {
		return EnrollDeviceResponse{}, errors.New("principal_id is required")
	}
	if workspaceID == "" {
		return EnrollDeviceResponse{}, errors.New("workspace_id is required")
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = "Riido Desktop"
	}
	secret, err := newDeviceSecret()
	if err != nil {
		return EnrollDeviceResponse{}, err
	}
	hash := sha256.Sum256([]byte(secret))
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextDeviceID++
	deviceID := fmt.Sprintf("device-enrolled-%06d", s.nextDeviceID)
	if s.deviceCredentials == nil {
		s.deviceCredentials = map[string]deviceCredentialRecord{}
	}
	s.deviceCredentials[deviceID] = deviceCredentialRecord{
		deviceID:         deviceID,
		secretHash:       hash,
		ownerPrincipalID: principal.PrincipalID,
		workspaceID:      workspaceID,
		displayName:      displayName,
		issuedAt:         now,
	}
	s.upsertEnrolledDeviceLocked(DeviceRecord{
		DeviceID:         deviceID,
		OwnerPrincipalID: principal.PrincipalID,
		DisplayName:      displayName,
		DaemonLastSeenAt: now,
	})
	if err := s.saveSnapshotLocked(ctx); err != nil {
		return EnrollDeviceResponse{}, err
	}
	return EnrollDeviceResponse{
		SchemaVersion:    DeviceCredentialSchemaVersion,
		DeviceID:         deviceID,
		DeviceSecret:     secret,
		OwnerPrincipalID: principal.PrincipalID,
		WorkspaceID:      workspaceID,
		DisplayName:      displayName,
		IssuedAt:         now,
	}, nil
}

func (s *DevelopmentAIAgentClientStore) AuthorizeDeviceCredential(ctx context.Context, deviceID, deviceSecret string, req AuthorizationRequest) (AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return AuthorizationResult{}, err
	}
	deviceID = strings.TrimSpace(deviceID)
	deviceSecret = strings.TrimSpace(deviceSecret)
	if deviceID == "" || deviceSecret == "" {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	if req.Resource != AuthorizationResourceAgent {
		return AuthorizationResult{}, ErrAuthorizationForbidden
	}

	s.mu.Lock()
	record, ok := s.deviceCredentials[deviceID]
	if !ok {
		s.mu.Unlock()
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	got := sha256.Sum256([]byte(deviceSecret))
	if subtle.ConstantTimeCompare(got[:], record.secretHash[:]) != 1 {
		s.mu.Unlock()
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	workspaceID := strings.TrimSpace(record.workspaceID)
	if requestedWorkspaceID := strings.TrimSpace(req.WorkspaceID); requestedWorkspaceID != "" {
		if workspaceID != "" && requestedWorkspaceID != workspaceID {
			s.mu.Unlock()
			return AuthorizationResult{}, ErrAuthorizationForbidden
		}
		workspaceID = requestedWorkspaceID
	}
	if agentID := strings.TrimSpace(req.AgentID); agentID != "" {
		agent, ok := s.agents[agentID]
		if !ok || s.agentWorkspaceID(agent) != workspaceID {
			s.mu.Unlock()
			return AuthorizationResult{}, ErrAuthorizationForbidden
		}
		binding, ok := s.agentRuntimeBindingLocked(agent)
		if !ok || binding.DeviceID != deviceID {
			s.mu.Unlock()
			return AuthorizationResult{}, ErrAuthorizationForbidden
		}
	}
	s.mu.Unlock()
	return AuthorizationResult{
		PrincipalID: record.ownerPrincipalID,
		WorkspaceID: workspaceID,
	}, nil
}

func (s *DevelopmentAIAgentClientStore) upsertEnrolledDeviceLocked(device DeviceRecord) {
	for i := range s.devices {
		if s.devices[i].DeviceID == device.DeviceID {
			if len(s.devices[i].Runtimes) > 0 && len(device.Runtimes) == 0 {
				device.Runtimes = copyRuntimes(s.devices[i].Runtimes)
			}
			s.devices[i] = device
			return
		}
	}
	s.devices = append(s.devices, device)
}

func newDeviceSecret() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate device secret: %w", err)
	}
	return "rdev_" + hex.EncodeToString(raw[:]), nil
}

func copyRuntimes(runtimes []RuntimeRecord) []RuntimeRecord {
	out := make([]RuntimeRecord, 0, len(runtimes))
	for _, runtime := range runtimes {
		out = append(out, copyRuntime(runtime))
	}
	return out
}
