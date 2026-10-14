// Package authpep adapts Riido Auth JWT/JWKS/introspection evidence to the
// control-plane RequestAuthorizer port. Domain authorization remains in the
// injected riidoaiserver authorizer.
package authpep

import (
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type (
	AuthorizationRequest = riidoaiserver.AuthorizationRequest
	AuthorizationResult  = riidoaiserver.AuthorizationResult
	RequestAuthorizer    = riidoaiserver.RequestAuthorizer
)

const (
	AuthorizationResourceAIAgentClient       = riidoaiserver.AuthorizationResourceAIAgentClient
	AuthorizationResourceAgent               = riidoaiserver.AuthorizationResourceAgent
	AuthorizationResourceAgentCatalog        = riidoaiserver.AuthorizationResourceAgentCatalog
	AuthorizationResourceComponentTask       = riidoaiserver.AuthorizationResourceComponentTask
	AuthorizationResourceComponentTaskEvents = riidoaiserver.AuthorizationResourceComponentTaskEvents
	AuthorizationResourceMetrics             = riidoaiserver.AuthorizationResourceMetrics

	AuthorizationActionAssign        = riidoaiserver.AuthorizationActionAssign
	AuthorizationActionCreate        = riidoaiserver.AuthorizationActionCreate
	AuthorizationActionDelete        = riidoaiserver.AuthorizationActionDelete
	AuthorizationActionDeviceControl = riidoaiserver.AuthorizationActionDeviceControl
	AuthorizationActionDeviceRead    = riidoaiserver.AuthorizationActionDeviceRead
	AuthorizationActionRead          = riidoaiserver.AuthorizationActionRead
	AuthorizationActionStream        = riidoaiserver.AuthorizationActionStream
	AuthorizationActionStop          = riidoaiserver.AuthorizationActionStop
	AuthorizationActionUpdate        = riidoaiserver.AuthorizationActionUpdate
)

var (
	ErrAuthorizationUnauthenticated   = riidoaiserver.ErrAuthorizationUnauthenticated
	ErrAuthorizationInvalidCredential = riidoaiserver.ErrAuthorizationInvalidCredential
	ErrAuthorizationForbidden         = riidoaiserver.ErrAuthorizationForbidden
)

func NewFallbackAuthorizer(authorizers ...RequestAuthorizer) (RequestAuthorizer, error) {
	return riidoaiserver.NewFallbackAuthorizer(authorizers...)
}

func validJWTEmail(value string) bool {
	parts := strings.Split(value, "@")
	return len(parts) == 2 && parts[0] != "" && strings.Contains(parts[1], ".") && len(value) <= 254 && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\r\n\x00 ")
}
