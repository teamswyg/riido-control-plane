package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func staticAuthorizerFromEnv(reviewProvision *riidoaiserver.ReviewAccountProvisioning) (riidoaiserver.RequestAuthorizer, error) {
	credentials, err := parseAuthzTokenCredentialsJSON(os.Getenv(envAuthzTokensJSON))
	if err != nil {
		return nil, err
	}
	if reviewProvision != nil {
		credentials = append(credentials, reviewProvision.Credential)
	}
	if len(credentials) == 0 {
		return nil, nil
	}
	authorizer, err := riidoaiserver.NewStaticTokenAuthorizer(credentials)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthzTokensJSON, err)
	}
	return authorizer, nil
}

func parseAuthzTokensJSON(raw string) (*riidoaiserver.StaticTokenAuthorizer, error) {
	credentials, err := parseAuthzTokenCredentialsJSON(raw)
	if err != nil {
		return nil, err
	}
	if len(credentials) == 0 {
		return nil, nil
	}
	return riidoaiserver.NewStaticTokenAuthorizer(credentials)
}

func parseAuthzTokenCredentialsJSON(raw string) ([]riidoaiserver.StaticTokenCredential, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var credentials []riidoaiserver.StaticTokenCredential
	if err := strictDecodeJSON(raw, &credentials); err != nil {
		return nil, fmt.Errorf("%s: %w", envAuthzTokensJSON, err)
	}
	return credentials, nil
}
