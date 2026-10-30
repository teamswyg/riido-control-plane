package riidoaiserver

import (
	"context"
	"errors"
	"time"
)

const (
	cliInstallationServicePrincipal = "service:riido-monolith"
	cliInstallationAlgorithm        = "p256-sha256-x962-der-v1"
	cliInstallationIntentSchema     = "riido-cli-installation-link-intent.v1"
	cliInstallationLinkSchema       = "riido-cli-installation-link.v1"
	cliInstallationPolicySchema     = "riido-cli-installation-policy.v1"
)

var (
	errCLIInstallationInvalid     = errors.New("invalid CLI installation request")
	errCLIInstallationConflict    = errors.New("CLI installation binding conflicts")
	errCLIInstallationExpired     = errors.New("CLI installation intent expired")
	errCLIInstallationUnavailable = errors.New("CLI installation unavailable")
)

type cliInstallationIdentity struct {
	InstallationID string `json:"installation_id"`
	Resource       string `json:"resource"`
	ClientID       string `json:"client_id"`
	KeyFingerprint string `json:"key_fingerprint"`
}

type cliInstallationIntentRequest struct {
	SchemaVersion string `json:"schema_version"`
	cliInstallationIdentity
	PublicKeyAlgorithm  string `json:"public_key_algorithm"`
	PublicKey           string `json:"public_key"`
	ExpectedDeviceID    string `json:"expected_device_id"`
	ApprovedPrincipalID string `json:"approved_principal_id"`
}

type cliInstallationIntent struct {
	Request         cliInstallationIntentRequest `json:"request"`
	ID              string                       `json:"id"`
	ChallengeDigest string                       `json:"challenge_digest"`
	ExpiresAt       time.Time                    `json:"expires_at"`
	Consumed        bool                         `json:"consumed"`
	Retired         bool                         `json:"retired,omitempty"`
	ReplayKey       string                       `json:"replay_key,omitempty"`
}

type cliInstallationLink struct {
	SchemaVersion string `json:"schema_version"`
	cliInstallationIdentity
	DeviceID          string    `json:"device_id"`
	LinkApprovedBy    string    `json:"link_approved_by"`
	DeviceOwnerAtLink string    `json:"device_owner_at_link"`
	LinkRevision      string    `json:"link_revision"`
	LinkedAt          time.Time `json:"linked_at"`
}

type cliInstallationRepository interface {
	createCLIInstallationIntent(context.Context, cliInstallationIntent) error
	createOrReplayCLIInstallationIntent(context.Context, cliInstallationIntentReplay) (cliInstallationIntentReplayResponse, error)
	confirmCLIInstallationLink(context.Context, string, authenticatedClaimDevice, string, string) (cliInstallationLink, error)
	readCLIInstallationPolicy(context.Context, cliInstallationIdentity, cliInstallationDeviceOwnerReader) (cliInstallationPolicy, error)
	readCLIInstallationReceipt(context.Context, cliInstallationReceiptRequest) (cliInstallationReceipt, error)
	retireCLIInstallationIntent(context.Context, cliInstallationRetirementRequest) (cliInstallationRetirement, error)
}

type cliInstallationDeviceOwnerReader interface {
	cliInstallationDeviceOwner(context.Context, string) (string, error)
}
