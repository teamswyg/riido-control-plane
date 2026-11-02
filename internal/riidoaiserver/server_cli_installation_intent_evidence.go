package riidoaiserver

import "net/http"

func (s Server) handleCLIInstallationIntentEvidence(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if err := s.authorizeCLIInstallationService(r, "link"); err != nil {
			return nil, err
		}
		var req cliInstallationIntentCreateRequest
		keys := "schema_version request_id installation_id resource client_id key_fingerprint public_key_algorithm public_key expected_device_id approved_principal_id"
		if err := decodeCLIInstallation(w, r, keys, &req); err != nil {
			return nil, err
		}
		if _, err := installationEvidenceLookup(req); err != nil {
			return nil, err
		}
		if req.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		store, ok := s.assignment.(cliInstallationIntentEvidenceReader)
		if !ok {
			return nil, errCLIInstallationUnavailable
		}
		return store.readCLIInstallationIntentEvidence(r.Context(), req)
	})
}
