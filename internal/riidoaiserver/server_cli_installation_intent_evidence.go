package riidoaiserver

import "net/http"

func (s Server) handleCLIInstallationIntentEvidence(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if err := s.authorizeCLIInstallationService(r, "link"); err != nil {
			return nil, err
		}
		req, err := decodeCLIInstallationIntentEvidence(w, r)
		if err != nil {
			return nil, err
		}
		lookup, err := installationEvidenceLookup(req)
		if err != nil {
			return nil, err
		}
		if lookup.Request.isTransfer() {
			if err := s.authorizeCLIInstallationService(r, "transfer"); err != nil {
				return nil, err
			}
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
