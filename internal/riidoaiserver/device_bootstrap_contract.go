package riidoaiserver

import (
	"context"
	"time"
)

const (
	deviceBootstrapSchema         = "riido-device-bootstrap.v1"
	deviceBootstrapFenceSchema    = "device_policy_bootstrap/v1"
	deviceBootstrapCoveragePrefix = "covb_"
)

// The Desktop persists the next secret in its encrypted journal BEFORE creating
// this request. Only its hash crosses this boundary; a lost ACK never requires
// the server to disclose or regenerate a native device credential.
type deviceBootstrapRequest struct {
	cliInstallationIntentRequest
	TransitionID   string `json:"transition_id"`
	NextSecretHash string `json:"next_secret_hash"`
}

type deviceBootstrapRecord struct {
	Schema              string                                `json:"schema"`
	Request             deviceBootstrapRequest                `json:"request"`
	Before              AIAgentClientDeviceCredentialSnapshot `json:"before"`
	After               AIAgentClientDeviceCredentialSnapshot `json:"after"`
	Intent              cliInstallationIntent                 `json:"intent"`
	State               string                                `json:"state"`
	Revision            uint64                                `json:"revision"`
	PolicyRevision      uint64                                `json:"policy_revision"`
	PreviousBootDigest  string                                `json:"previous_boot_digest,omitempty"`
	CompletedBootDigest string                                `json:"completed_boot_digest,omitempty"`
	CompletedAt         time.Time                             `json:"completed_at,omitempty"`
}

type deviceBootstrapProof struct {
	IntentID   string `json:"intent_id"`
	BootDigest string `json:"boot_digest"`
	Signature  string `json:"signature"`
}

type deviceBootstrapRepository interface {
	readDeviceBootstrap(context.Context, string) (deviceBootstrapRecord, bool, error)
	readBootstrapLegacyRevision(context.Context, string) (uint64, error)
	verifyBootstrapState(context.Context, deviceBootstrapRecord) error
}

type deviceBootstrapMutation struct {
	before *deviceBootstrapRecord
	after  deviceBootstrapRecord
}
type deviceBootstrapMutationKey struct{}

// Constructed only after native device credential authorization.
type deviceBootstrapAuth struct {
	deviceID, ownerPrincipalID, executionCoverageGeneration, credentialGeneration string
}
