package riidoaiserver

import (
	"net/http"
)

func (s Server) handleDeviceBootstrapOperation(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if !cliInstallationResource(s.config.CLIInstallationService.Resource) {
			return nil, errCLIInstallationUnavailable
		}
		authenticated, err := s.authorizeCLIInstallationDevice(r)
		if err != nil {
			return nil, err
		}
		auth := deviceBootstrapAuth{authenticated.deviceID, authenticated.ownerPrincipalID, authenticated.coverage.generation, authenticated.coverage.credential}
		store, ok := s.devices.(deviceBootstrapStore)
		if !ok {
			return nil, errCLIInstallationUnavailable
		}
		id, operation := r.PathValue("transition_id"), r.PathValue("operation")
		keys := "schema_version"
		switch operation {
		case "prepare", "activate":
			keys += " intent_id boot_digest signature"
		case "status", "renew":
		default:
			return nil, errCLIInstallationInvalid
		}
		var request struct {
			SchemaVersion string `json:"schema_version"`
			deviceBootstrapProof
		}
		if err := decodeCLIInstallation(w, r, keys, &request); err != nil {
			return nil, err
		}
		if request.SchemaVersion != deviceBootstrapSchema {
			return nil, errCLIInstallationInvalid
		}
		record, err := store.loadDeviceBootstrap(r.Context(), id, auth)
		if err != nil {
			return nil, err
		}
		if record.Request.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		switch operation {
		case "renew":
			record, err = store.renewDeviceBootstrap(r.Context(), id, auth)
		case "prepare", "activate":
			record, err = store.advanceDeviceBootstrap(r.Context(), id, operation, auth, request.deviceBootstrapProof)
		}
		if err != nil {
			return nil, err
		}
		return bootstrapResponse(record), nil
	})
}
