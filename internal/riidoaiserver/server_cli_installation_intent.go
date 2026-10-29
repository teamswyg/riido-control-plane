package riidoaiserver

import (
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
		wire, err := decodeCLIInstallationIntentCreate(w, r)
		if err != nil {
			return nil, err
		}
		req, requestID, err := normalizeCLIInstallationIntentCreate(wire)
		if err != nil {
			return nil, err
		}
		if req.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		store, err := s.cliInstallationRepository()
		if err != nil {
			return nil, err
		}
		if requestID != "" {
			record, err := newCLIInstallationIntentReplay(req, requestID)
			if err != nil {
				return nil, err
			}
			return store.createOrReplayCLIInstallationIntent(r.Context(), record)
		}
		intent, response, err := newCLIInstallationIntent(req)
		if err != nil {
			return nil, err
		}
		if err = store.createCLIInstallationIntent(r.Context(), intent); err != nil {
			return nil, err
		}
		return response, nil
	})
}
