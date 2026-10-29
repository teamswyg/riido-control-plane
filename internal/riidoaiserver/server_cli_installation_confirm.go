package riidoaiserver

import "net/http"

func (s Server) handleCLIInstallationConfirm(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		device, err := s.authorizeCLIInstallationDevice(r)
		if err != nil {
			return nil, err
		}
		var req struct {
			SchemaVersion string `json:"schema_version"`
			Signature     string `json:"signature"`
		}
		if err = decodeCLIInstallation(w, r, "schema_version signature", &req); err != nil {
			return nil, err
		}
		id := r.PathValue("intent_id")
		if req.SchemaVersion != "riido-cli-installation-link-confirm.v1" || !cliInstallationID(id) || len(req.Signature) > 128 {
			return nil, errCLIInstallationInvalid
		}
		store, err := s.cliInstallationRepository()
		if err != nil {
			return nil, err
		}
		if !cliInstallationResource(s.config.CLIInstallationService.Resource) {
			return nil, errCLIInstallationUnavailable
		}
		return store.confirmCLIInstallationLink(r.Context(), id, device, req.Signature, s.config.CLIInstallationService.Resource)
	})
}
