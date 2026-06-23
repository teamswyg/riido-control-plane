package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"time"
)

const DeviceCredentialSchemaVersion = "riido-device-credential.v1"

type EnrollDeviceRequest struct {
	DeviceID string `json:"device_id,omitempty"`
	// MachineID is unique and stable to the physical device. When present, the
	// DeviceID is derived from the machine so every workspace resolves to one row.
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
