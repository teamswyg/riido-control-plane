package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func newExternalAuthorizer(endpoint, apiKey string) (riidoaiserver.RequestAuthorizer, error) {
	timeout, err := envDurationSeconds(envExternalAuthzTimeout, 0)
	if err != nil {
		return nil, err
	}
	authorizer, err := riidoaiserver.NewExternalHTTPAuthorizer(riidoaiserver.ExternalHTTPAuthorizerConfig{
		Endpoint: endpoint,
		Audience: strings.TrimSpace(os.Getenv(envExternalAuthzAudience)),
		APIKey:   apiKey,
		Timeout:  timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envExternalAuthzURL, err)
	}
	return riidoaiserver.NewCoalescingAuthorizer(authorizer)
}
