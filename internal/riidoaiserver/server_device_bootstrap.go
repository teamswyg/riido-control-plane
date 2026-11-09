package riidoaiserver

import (
	"context"
	"net/http"
)

type deviceBootstrapStore interface {
	beginDeviceBootstrap(context.Context, deviceBootstrapRequest) (deviceBootstrapRecord, error)
	loadDeviceBootstrap(context.Context, string, deviceBootstrapAuth) (deviceBootstrapRecord, error)
	renewDeviceBootstrap(context.Context, string, deviceBootstrapAuth) (deviceBootstrapRecord, error)
	advanceDeviceBootstrap(context.Context, string, string, deviceBootstrapAuth, deviceBootstrapProof) (deviceBootstrapRecord, error)
}

func (s Server) handleDeviceBootstrapCreate(w http.ResponseWriter, r *http.Request) {
	installationHTTP(w, r, func(r *http.Request) (any, error) {
		if err := s.authorizeCLIInstallationService(r, "bootstrap"); err != nil {
			return nil, err
		}
		var request deviceBootstrapRequest
		if err := decodeCLIInstallation(w, r, "schema_version resource client_id installation_id key_fingerprint public_key_algorithm public_key expected_device_id approved_principal_id transition_id next_secret_hash", &request); err != nil {
			return nil, err
		}
		if request.SchemaVersion != deviceBootstrapSchema {
			return nil, errCLIInstallationInvalid
		}
		request.SchemaVersion = cliInstallationIntentSchema
		if request.Resource != s.config.CLIInstallationService.Resource {
			return nil, ErrAuthorizationForbidden
		}
		store, ok := s.devices.(deviceBootstrapStore)
		if !ok {
			return nil, errCLIInstallationUnavailable
		}
		record, err := store.beginDeviceBootstrap(r.Context(), request)
		if err != nil {
			return nil, err
		}
		return bootstrapResponse(record), nil
	})
}
