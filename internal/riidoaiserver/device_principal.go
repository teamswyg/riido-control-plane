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
	DeviceID string `json:"device_id,omitempty"`
	// MachineID is a value unique and stable to the physical device (the daemon's
	// persistent machine UUID). When present, the DeviceID is derived
	// deterministically from (principal, machine) so the same machine resolves to
	// exactly one device row regardless of how many workspaces it enrolls in.
	MachineID   string `json:"machine_id,omitempty"`
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
	machineID        string
	secretHash       [sha256.Size]byte
	ownerPrincipalID string
	workspaceID      string
	displayName      string
	issuedAt         time.Time
}

// deviceIDForMachine derives a stable DeviceID from the device's unique machine
// id ALONE — not the principal. One physical machine maps to exactly one device
// row, shared across the accounts/workspaces it connects to (a device is an
// entity of the machine, not of an account). The DeviceID is not a secret; the
// rotating device secret is the auth factor.
func deviceIDForMachine(machineID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(machineID)))
	return "dev_" + hex.EncodeToString(sum[:16])
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

	machineID := strings.TrimSpace(req.MachineID)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.deviceCredentials == nil {
		s.deviceCredentials = map[string]deviceCredentialRecord{}
	}
	var deviceID string
	switch {
	case machineID != "":
		// One device per machine, shared across the workspaces it connects to.
		// Derive the DeviceID from the machine alone so every enrollment of this
		// machine resolves to the same device row. Only the first enrolling
		// principal holds the daemon credential; other accounts gain visibility
		// through workspace connection, not by re-enrolling.
		deviceID = deviceIDForMachine(machineID)
		if existing, ok := s.deviceCredentials[deviceID]; ok && existing.ownerPrincipalID != principal.PrincipalID {
			return EnrollDeviceResponse{}, ErrAuthorizationForbidden
		}
	case strings.TrimSpace(req.DeviceID) != "":
		// Legacy reuse path: a caller without machine_id may reuse a prior DeviceID
		// only within the same (principal, workspace) it was issued for.
		deviceID = strings.TrimSpace(req.DeviceID)
		existing, ok := s.deviceCredentials[deviceID]
		if !ok ||
			existing.ownerPrincipalID != principal.PrincipalID ||
			existing.workspaceID != workspaceID {
			return EnrollDeviceResponse{}, ErrAuthorizationForbidden
		}
	default:
		s.nextDeviceCredentialSeq++
		deviceID = fmt.Sprintf("device-enrolled-%06d", s.nextDeviceCredentialSeq)
	}
	s.deviceCredentials[deviceID] = deviceCredentialRecord{
		deviceID:         deviceID,
		machineID:        machineID,
		secretHash:       hash,
		ownerPrincipalID: principal.PrincipalID,
		workspaceID:      workspaceID,
		displayName:      displayName,
		issuedAt:         now,
	}
	s.upsertEnrolledDeviceLocked(DeviceRecord{
		DeviceID:              deviceID,
		OwnerPrincipalID:      principal.PrincipalID,
		DisplayName:           displayName,
		DaemonLastSeenAt:      now,
		ConnectedWorkspaceIDs: []string{workspaceID},
	})
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

// legacyDaemonIDPrefix is the hardcoded daemon id used before per-machine UUIDs.
// Runtime IDs minted under it (agentd-local:<provider>) were identical on every
// machine, so the control-plane's move-to-last-reporter rule shuffled them
// between device rows and left devices showing no runtimes. They are pure stale
// residue and must not survive into the per-machine model.
const legacyDaemonIDPrefix = "agentd-local:"

// pruneLegacyRuntimeRecords drops legacy globally-colliding runtime records from
// restored device rows. Device rows themselves are preserved (they may simply be
// offline); only the unusable legacy runtimes are removed.
func pruneLegacyRuntimeRecords(devices []DeviceRecord) []DeviceRecord {
	for i := range devices {
		if len(devices[i].Runtimes) == 0 {
			continue
		}
		kept := devices[i].Runtimes[:0]
		for _, runtime := range devices[i].Runtimes {
			if strings.HasPrefix(strings.TrimSpace(runtime.RuntimeID), legacyDaemonIDPrefix) {
				continue
			}
			kept = append(kept, runtime)
		}
		devices[i].Runtimes = kept
	}
	return devices
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
	s.mu.Unlock()
	if !ok {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	got := sha256.Sum256([]byte(deviceSecret))
	if subtle.ConstantTimeCompare(got[:], record.secretHash[:]) != 1 {
		return AuthorizationResult{}, ErrAuthorizationUnauthenticated
	}
	workspaceID := strings.TrimSpace(req.WorkspaceID)
	if workspaceID == "" {
		workspaceID = record.workspaceID
	}
	return AuthorizationResult{
		PrincipalID: record.ownerPrincipalID,
		WorkspaceID: workspaceID,
	}, nil
}

func (s *DevelopmentAIAgentClientStore) upsertEnrolledDeviceLocked(device DeviceRecord) {
	for i := range s.devices {
		if s.devices[i].DeviceID == device.DeviceID {
			// Re-enroll (idempotent machine path) must not wipe runtimes already
			// detected for this device — only refresh ownership/name/last-seen.
			merged := copyDevice(s.devices[i])
			merged.OwnerPrincipalID = device.OwnerPrincipalID
			if device.DisplayName != "" {
				merged.DisplayName = device.DisplayName
			}
			if !device.DaemonLastSeenAt.IsZero() {
				merged.DaemonLastSeenAt = device.DaemonLastSeenAt
			}
			for _, ws := range device.ConnectedWorkspaceIDs {
				merged.ConnectedWorkspaceIDs = addConnectedWorkspace(merged.ConnectedWorkspaceIDs, ws)
			}
			s.devices[i] = merged
			return
		}
	}
	s.devices = append(s.devices, copyDevice(device))
}

func newDeviceSecret() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate device secret: %w", err)
	}
	return "rdev_" + hex.EncodeToString(raw[:]), nil
}
