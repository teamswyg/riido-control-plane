package authpep

import (
	"context"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

// RestrictLegacyApplicationTokens keeps monolith CLI grants out of the app
// authorizer. That PDP does not enforce CLI read scopes, live grant revocation
// or installation policy. Classification can only deny; it never authenticates
// a token. CLI credentials require a dedicated resource-bound read path.
func RestrictLegacyApplicationTokens(next riidoaiserver.RequestAuthorizer) riidoaiserver.RequestAuthorizer {
	if next == nil {
		return nil
	}
	return &legacyApplicationAuthorizer{next: next}
}

type legacyApplicationAuthorizer struct {
	next riidoaiserver.RequestAuthorizer
}

func (a *legacyApplicationAuthorizer) Authorize(ctx context.Context, raw string, req riidoaiserver.AuthorizationRequest) (riidoaiserver.AuthorizationResult, error) {
	if err := ctx.Err(); err != nil {
		return riidoaiserver.AuthorizationResult{}, err
	}
	if rejectLegacyCLIToken(raw) {
		return riidoaiserver.AuthorizationResult{}, riidoaiserver.ErrAuthorizationForbidden
	}
	return a.next.Authorize(ctx, raw, req)
}
