package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) authorizeCLIInstallationService(r *http.Request, action string) error {
	cfg := s.config.CLIInstallationService
	if cfg.authorizer == nil || !cliInstallationResource(cfg.Resource) {
		return errCLIInstallationUnavailable
	}
	values := r.Header.Values("Authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer ") || len(values[0]) > 8192 || len(r.Header.Values(deviceIDHeader)) > 0 || len(r.Header.Values(deviceSecretHeader)) > 0 || len(r.Header.Values("X-Riido-Ai-Agent-Token")) > 0 {
		return ErrAuthorizationUnauthenticated
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	if token == "" || strings.ContainsAny(token, " \t\r\n") {
		return ErrAuthorizationUnauthenticated
	}
	principal, err := cfg.authorizer.Authorize(r.Context(), token, AuthorizationRequest{Resource: "cli_installation", Action: AuthorizationAction(action)})
	if err != nil {
		return err
	}
	if principal.PrincipalID != cliInstallationServicePrincipal {
		return ErrAuthorizationForbidden
	}
	return nil
}

func (s Server) authorizeCLIInstallationDevice(r *http.Request) (authenticatedClaimDevice, error) {
	id, validID := singleExecutionHeader(r, deviceIDHeader, 512)
	secret, validSecret := singleExecutionHeader(r, deviceSecretHeader, 8192)
	if !validID || !validSecret || len(r.Header.Values("Authorization")) > 0 || len(r.Header.Values("X-Riido-Ai-Agent-Token")) > 0 {
		return authenticatedClaimDevice{}, ErrAuthorizationUnauthenticated
	}
	if s.devices == nil {
		return authenticatedClaimDevice{}, errCLIInstallationUnavailable
	}
	principal, err := s.devices.AuthorizeDeviceCredential(r.Context(), id, secret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, DeviceID: id})
	if err != nil {
		return authenticatedClaimDevice{}, err
	}
	if !cliInstallationID(principal.PrincipalID) {
		return authenticatedClaimDevice{}, ErrAuthorizationForbidden
	}
	return authenticatedClaimDevice{id, principal.PrincipalID}, nil
}
