package riidoaiserver

import (
	"net/http"
	"strings"
)

func (s Server) authorizeExecutionTermination(r *http.Request) (executionTerminationIdentity, error) {
	device, validDevice := singleExecutionHeader(r, deviceIDHeader, 512)
	secret, validSecret := singleExecutionHeader(r, deviceSecretHeader, 8192)
	lease, validLease := singleExecutionHeader(r, executionLeaseHeader, 8192)
	if !validDevice || !validSecret || !validLease {
		return executionTerminationIdentity{}, ErrAuthorizationUnauthenticated
	}
	assignmentID := r.PathValue("assignment_id")
	if assignmentID == "" || len(assignmentID) > 512 || strings.TrimSpace(assignmentID) != assignmentID || strings.ContainsAny(assignmentID, "/\\\x00") {
		return executionTerminationIdentity{}, errTerminationForbidden
	}
	if s.devices == nil {
		return executionTerminationIdentity{}, errTerminationUnavailable
	}
	principal, err := s.devices.AuthorizeDeviceCredential(r.Context(), device, secret, AuthorizationRequest{Resource: AuthorizationResourceAgent, Action: AuthorizationActionEventsWrite, DeviceID: device})
	if err != nil {
		return executionTerminationIdentity{}, err
	}
	if strings.TrimSpace(principal.PrincipalID) == "" {
		return executionTerminationIdentity{}, errTerminationForbidden
	}
	return executionTerminationIdentity{deviceID: device, ownerPrincipalID: principal.PrincipalID, assignmentID: assignmentID, epoch: leaseEpoch(lease), coverage: executionCoverageAuth{principal.executionCoverageGeneration, principal.credentialGeneration}}, nil
}

func singleExecutionHeader(r *http.Request, name string, limit int) (string, bool) {
	values := r.Header.Values(name)
	if len(values) != 1 || values[0] == "" || len(values[0]) > limit {
		return "", false
	}
	for _, value := range values[0] {
		if value < 33 || value > 126 {
			return "", false
		}
	}
	return values[0], true
}
