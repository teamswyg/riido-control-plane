package authpep

import (
	"errors"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

// IssuerAuthorizationSchemeSelector owns only credential classification. It
// does not trust claims or grant access; the selected PEP performs verification.
type IssuerAuthorizationSchemeSelector struct {
	issuer string
}

func NewIssuerAuthorizationSchemeSelector(issuer string) (*IssuerAuthorizationSchemeSelector, error) {
	issuer = strings.TrimSpace(issuer)
	if !exactHTTPSOrigin(issuer) {
		return nil, errors.New("exact HTTPS Auth issuer is required")
	}
	return &IssuerAuthorizationSchemeSelector{issuer: issuer}, nil
}

func (s *IssuerAuthorizationSchemeSelector) SelectAuthorizationScheme(raw string) (riidoaiserver.AuthorizationScheme, error) {
	if s != nil && claimsConfiguredIssuer(raw, s.issuer) {
		return riidoaiserver.AuthorizationSchemeAuthServiceV2, nil
	}
	return riidoaiserver.AuthorizationSchemeLegacyV1, nil
}
