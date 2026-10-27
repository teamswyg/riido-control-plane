package riidoaiserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"
)

type cliInstallationIntentResponse struct {
	SchemaVersion string    `json:"schema_version"`
	IntentID      string    `json:"intent_id"`
	Challenge     string    `json:"challenge"`
	ExpiresAt     time.Time `json:"expires_at"`
}

func (s Server) handleCLIInstallationIntent(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if err := s.authorizeCLIInstallationService(r, "link"); err != nil {
			return nil, err
		}
		var req cliInstallationIntentRequest
		keys := "schema_version installation_id resource client_id key_fingerprint public_key_algorithm public_key expected_device_id approved_principal_id"
		if err := decodeCLIInstallation(w, r, keys, &req); err != nil {
			return nil, err
		}
		if req.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		if _, err := installationPublicKey(req); err != nil {
			return nil, err
		}
		store, err := s.cliInstallationRepository()
		if err != nil {
			return nil, err
		}
		var random [64]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, errCLIInstallationUnavailable
		}
		challenge := base64.RawURLEncoding.EncodeToString(random[32:])
		digest := sha256.Sum256(random[32:])
		intent := cliInstallationIntent{Request: req, ID: base64.RawURLEncoding.EncodeToString(random[:32]), ChallengeDigest: hex.EncodeToString(digest[:]), ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
		if err = store.createCLIInstallationIntent(r.Context(), intent); err != nil {
			return nil, err
		}
		return cliInstallationIntentResponse{cliInstallationIntentSchema, intent.ID, challenge, intent.ExpiresAt}, nil
	})
}
