package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func authorizerFromEnv() (riidoaiserver.RequestAuthorizer, error) {
	reviewProvision, err := reviewAccountProvisioningFromEnv()
	if err != nil {
		return nil, err
	}
	return authorizerFromEnvWithReview(reviewProvision)
}

func authorizerFromEnvWithReview(reviewProvision *riidoaiserver.ReviewAccountProvisioning) (riidoaiserver.RequestAuthorizer, error) {
	static, err := staticAuthorizerFromEnv(reviewProvision)
	if err != nil {
		return nil, err
	}
	external, err := externalAuthorizerFromEnv()
	if err != nil {
		return nil, err
	}
	jwt, err := jwtAuthorizerFromEnv(external)
	if err != nil {
		return nil, err
	}
	var legacy []riidoaiserver.RequestAuthorizer
	if static != nil {
		legacy = append(legacy, static)
	}
	if external != nil {
		legacy = append(legacy, external)
	}
	legacyAuthorizer, err := composeAuthorizers(legacy)
	if err != nil {
		return nil, err
	}
	return composeAuthorizationSchemes(strings.TrimSpace(os.Getenv(envAuthIssuer)), legacyAuthorizer, jwt)
}

func composeAuthorizers(authorizers []riidoaiserver.RequestAuthorizer) (riidoaiserver.RequestAuthorizer, error) {
	switch len(authorizers) {
	case 0:
		return nil, nil
	case 1:
		return authorizers[0], nil
	default:
		return riidoaiserver.NewFallbackAuthorizer(authorizers...)
	}
}

func externalAuthorizerFromEnv() (riidoaiserver.RequestAuthorizer, error) {
	endpoint := strings.TrimSpace(os.Getenv(envExternalAuthzURL))
	apiKey := strings.TrimSpace(os.Getenv(envExternalAuthzAPIKey))
	if endpoint == "" && apiKey != "" {
		return nil, fmt.Errorf("%s requires %s", envExternalAuthzAPIKey, envExternalAuthzURL)
	}
	if endpoint == "" {
		return nil, nil
	}
	return newExternalAuthorizer(endpoint, apiKey)
}
