package riidoaiserver

import (
	"net/http"
	"time"
)

type cliInstallationPolicy struct {
	SchemaVersion     string    `json:"schema_version"`
	Decision          string    `json:"decision"`
	LinkRevision      string    `json:"link_revision"`
	ExecutionRevision string    `json:"execution_revision"`
	CoverageState     string    `json:"coverage_state"`
	CoverageID        string    `json:"coverage_id"`
	CheckedAt         time.Time `json:"checked_at"`
}

func (s Server) handleCLIInstallationPolicy(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if err := s.authorizeCLIInstallationService(r, "policy-read"); err != nil {
			return nil, err
		}
		var req struct {
			SchemaVersion string `json:"schema_version"`
			cliInstallationIdentity
		}
		if err := decodeCLIInstallation(w, r, "schema_version installation_id resource client_id key_fingerprint", &req); err != nil {
			return nil, err
		}
		if req.SchemaVersion != cliInstallationPolicySchema || !req.valid() {
			return nil, errCLIInstallationInvalid
		}
		if req.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		store, err := s.cliInstallationRepository()
		if err != nil {
			return nil, err
		}
		owner, ok := s.devices.(cliInstallationDeviceOwnerReader)
		if !ok {
			return nil, errCLIInstallationUnavailable
		}
		return store.readCLIInstallationPolicy(r.Context(), req.cliInstallationIdentity, owner)
	})
}
