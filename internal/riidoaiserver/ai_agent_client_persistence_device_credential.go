package riidoaiserver

import "time"

type AIAgentClientDeviceCredentialSnapshot struct {
	DeviceID         string    `json:"device_id"`
	MachineID        string    `json:"machine_id,omitempty"`
	SecretHashSHA256 string    `json:"secret_hash_sha256"`
	OwnerPrincipalID string    `json:"owner_principal_id"`
	WorkspaceID      string    `json:"workspace_id"`
	DisplayName      string    `json:"display_name,omitempty"`
	IssuedAt         time.Time `json:"issued_at"`
}
