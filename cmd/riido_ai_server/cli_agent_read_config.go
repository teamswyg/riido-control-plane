package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/authpep"
	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const (
	envCLIAgentReadEnabled        = "RIIDO_AI_SERVER_CLI_AGENT_READ_ENABLED"
	envCLIAgentReadMonolithOrigin = "RIIDO_AI_SERVER_CLI_AGENT_READ_MONOLITH_ORIGIN"
	envCLIAgentReadResource       = "RIIDO_AI_SERVER_CLI_AGENT_READ_RESOURCE"
	envCLIAgentReadServiceToken   = "RIIDO_AI_SERVER_CLI_AGENT_READ_SERVICE_TOKEN"
)

func cliAgentReadAuthorizerFromEnv() (riidoaiserver.RequestAuthorizer, error) {
	switch os.Getenv(envCLIAgentReadEnabled) {
	case "", "false":
		return nil, nil
	case "true":
	default:
		return nil, fmt.Errorf("%s must be true or false", envCLIAgentReadEnabled)
	}
	token := os.Getenv(envCLIAgentReadServiceToken)
	if token == strings.TrimSpace(os.Getenv(envExternalAuthzAPIKey)) || token == strings.TrimSpace(os.Getenv(envCLIThreadReadServiceToken)) {
		return nil, fmt.Errorf("CLI agent read requires a dedicated service credential")
	}
	credentials, err := parseAuthzTokenCredentialsJSON(os.Getenv(envAuthzTokensJSON))
	if err != nil {
		return nil, err
	}
	reviewProvision, err := reviewAccountProvisioningFromEnv()
	if err != nil {
		return nil, err
	}
	if reviewProvision != nil {
		credentials = append(credentials, reviewProvision.Credential)
	}
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	for _, credential := range credentials {
		credential.Token = strings.TrimSpace(credential.Token)
		credential.TokenSHA256 = strings.TrimSpace(credential.TokenSHA256)
		if token == credential.Token || strings.EqualFold(tokenHash, credential.TokenSHA256) {
			return nil, fmt.Errorf("CLI agent read cannot reuse a static service credential")
		}
	}
	authorizer, err := authpep.NewCLIAgentReadAuthorizer(authpep.CLIThreadReadConfig{MonolithOrigin: os.Getenv(envCLIAgentReadMonolithOrigin), Resource: os.Getenv(envCLIAgentReadResource), ServiceToken: token})
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

func runtimeCLIAgentRead(config runtimeConfig) (runtimeConfig, error) {
	authorizer, err := cliAgentReadAuthorizerFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	config.CLIAgentReadAuthorizer = authorizer
	return config, nil
}
