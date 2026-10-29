package riidoaiserver

import "time"

type AIAgentClientDeviceCredentialSnapshot struct {
	ExecutionCoverageGeneration string    `json:"execution_coverage_generation,omitempty"`
	CredentialGeneration        string    `json:"credential_generation,omitempty"`
	DeviceID                    string    `json:"device_id"`
	MachineID                   string    `json:"machine_id,omitempty"`
	SecretHashSHA256            string    `json:"secret_hash_sha256"`
	OwnerPrincipalID            string    `json:"owner_principal_id"`
	WorkspaceID                 string    `json:"workspace_id"`
	DisplayName                 string    `json:"display_name,omitempty"`
	IssuedAt                    time.Time `json:"issued_at"`
}
