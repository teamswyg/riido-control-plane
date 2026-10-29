package riidoaiserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

func newCLIInstallationIntent(req cliInstallationIntentRequest) (cliInstallationIntent, cliInstallationIntentResponse, error) {
	var random [64]byte
	if _, err := rand.Read(random[:]); err != nil {
		return cliInstallationIntent{}, cliInstallationIntentResponse{}, errCLIInstallationUnavailable
	}
	challenge := base64.RawURLEncoding.EncodeToString(random[32:])
	digest := sha256.Sum256(random[32:])
	intent := cliInstallationIntent{Request: req, ID: base64.RawURLEncoding.EncodeToString(random[:32]), ChallengeDigest: hex.EncodeToString(digest[:]), ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	response := cliInstallationIntentResponse{cliInstallationIntentSchema, intent.ID, challenge, intent.ExpiresAt}
	return intent, response, nil
}

func newCLIInstallationIntentReplay(req cliInstallationIntentRequest, requestID string) (cliInstallationIntentReplay, error) {
	var challenge [32]byte
	if _, err := rand.Read(challenge[:]); err != nil {
		return cliInstallationIntentReplay{}, errCLIInstallationUnavailable
	}
	req.SchemaVersion = cliInstallationIntentV2Schema
	response := cliInstallationIntentResponse{SchemaVersion: cliInstallationIntentV2Schema, Challenge: base64.RawURLEncoding.EncodeToString(challenge[:]), ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	record := cliInstallationIntentReplay{req, cliInstallationIntentReplayResponse{response, requestID}}
	record.Response.IntentID = record.intentID()
	return record, nil
}
