package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/internal/authpep"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func composeAuthorizationSchemes(issuer string, legacy, v2 riidoaiserver.RequestAuthorizer) (riidoaiserver.RequestAuthorizer, error) {
	if v2 == nil {
		return legacy, nil
	}
	if legacy == nil {
		return nil, fmt.Errorf("auth_service_v2 requires the legacy_v1 boundary during migration")
	}
	selector, err := authpep.NewIssuerAuthorizationSchemeSelector(issuer)
	if err != nil {
		return nil, err
	}
	return riidoaiserver.NewAuthorizationSchemeRouter(selector, legacy, v2)
}
