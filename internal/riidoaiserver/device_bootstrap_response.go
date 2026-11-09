package riidoaiserver

import "time"

// Never serialize the persistence record at an HTTP boundary: it contains
// native credential hashes, the approval key and the immutable source snapshot.
type deviceBootstrapResponse struct {
	SchemaVersion string `json:"schema_version"`
	cliInstallationIdentity
	TransitionID                string     `json:"transition_id"`
	DeviceID                    string     `json:"device_id"`
	OwnerPrincipalID            string     `json:"owner_principal_id"`
	State                       string     `json:"state"`
	IntentID                    string     `json:"intent_id"`
	ChallengeDigest             string     `json:"challenge_digest"`
	ExpiresAt                   time.Time  `json:"expires_at"`
	PreviousBootDigest          string     `json:"previous_boot_digest,omitempty"`
	ExecutionCoverageGeneration string     `json:"execution_coverage_generation,omitempty"`
	CredentialGeneration        string     `json:"credential_generation,omitempty"`
	CompletedAt                 *time.Time `json:"completed_at,omitempty"`
}

func bootstrapResponse(r deviceBootstrapRecord) deviceBootstrapResponse {
	response := deviceBootstrapResponse{SchemaVersion: deviceBootstrapSchema, cliInstallationIdentity: r.Request.cliInstallationIdentity, TransitionID: r.Request.TransitionID, DeviceID: r.Before.DeviceID, OwnerPrincipalID: r.Before.OwnerPrincipalID, State: r.State, IntentID: r.Intent.ID, ChallengeDigest: r.Intent.ChallengeDigest, ExpiresAt: r.Intent.ExpiresAt, PreviousBootDigest: r.PreviousBootDigest}
	if r.State == "complete" {
		response.ExecutionCoverageGeneration = r.After.ExecutionCoverageGeneration
		response.CredentialGeneration = r.After.CredentialGeneration
		response.CompletedAt = &r.CompletedAt
	}
	return response
}
