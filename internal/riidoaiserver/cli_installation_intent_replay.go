package riidoaiserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
)

const cliInstallationIntentV2Schema = "riido-cli-installation-link-intent.v2"

type cliInstallationIntentCreateRequest struct {
	cliInstallationIntentRequest
	RequestID json.RawMessage `json:"request_id"`
}

type cliInstallationIntentReplayResponse struct {
	cliInstallationIntentResponse
	RequestID string `json:"request_id"`
}

type cliInstallationIntentReplay struct {
	Request  cliInstallationIntentRequest        `json:"request"`
	Response cliInstallationIntentReplayResponse `json:"response"`
}

func (record cliInstallationIntentReplay) key() string {
	digest := installationIntentRequestDigest("riido-cli-installation-intent-request.v1", record.Request, record.Response.RequestID)
	return hex.EncodeToString(digest[:])
}

func (record cliInstallationIntentReplay) intentID() string {
	digest := installationIntentRequestDigest("riido-cli-installation-intent-id.v2", record.Request, record.Response.RequestID)
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func installationRandom32(value string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(raw) == 32 && base64.RawURLEncoding.EncodeToString(raw) == value
}

func (record cliInstallationIntentReplay) intent() cliInstallationIntent {
	raw, _ := base64.RawURLEncoding.DecodeString(record.Response.Challenge)
	digest := sha256.Sum256(raw)
	request := record.Request
	request.SchemaVersion = cliInstallationIntentSchema
	return cliInstallationIntent{Request: request, ID: record.Response.IntentID, ChallengeDigest: hex.EncodeToString(digest[:]), ExpiresAt: record.Response.ExpiresAt, ReplayKey: record.key()}
}

func (record cliInstallationIntentReplay) matches(intent cliInstallationIntent) bool {
	expected := record.intent()
	return expected.Request == intent.Request && expected.ID == intent.ID && expected.ChallengeDigest == intent.ChallengeDigest && expected.ExpiresAt.Equal(intent.ExpiresAt) && expected.ReplayKey == intent.ReplayKey
}
