package riidoaiserver

import "net/http"

func (s Server) handleCLIInstallationRetirement(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if err := s.authorizeCLIInstallationService(r, "link"); err != nil {
			return nil, err
		}
		var req cliInstallationRetirementRequest
		keys := "schema_version intent_id installation_id resource client_id key_fingerprint expected_device_id approved_principal_id"
		if err := decodeCLIInstallation(w, r, keys, &req); err != nil {
			return nil, err
		}
		if !req.valid() {
			return nil, errCLIInstallationInvalid
		}
		if req.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		store, err := s.cliInstallationRepository()
		if err != nil {
			return nil, err
		}
		return store.retireCLIInstallationIntent(r.Context(), req)
	})
}
