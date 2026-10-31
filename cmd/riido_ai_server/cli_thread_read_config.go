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
	envCLIThreadReadEnabled        = "RIIDO_AI_SERVER_CLI_THREAD_READ_ENABLED"
	envCLIThreadReadMonolithOrigin = "RIIDO_AI_SERVER_CLI_THREAD_READ_MONOLITH_ORIGIN"
	envCLIThreadReadResource       = "RIIDO_AI_SERVER_CLI_THREAD_READ_RESOURCE"
	envCLIThreadReadServiceToken   = "RIIDO_AI_SERVER_CLI_THREAD_READ_SERVICE_TOKEN"
)

func cliThreadReadAuthorizerFromEnv() (riidoaiserver.RequestAuthorizer, error) {
	switch os.Getenv(envCLIThreadReadEnabled) {
	case "", "false":
		return nil, nil
	case "true":
	default:
		return nil, fmt.Errorf("%s must be true or false", envCLIThreadReadEnabled)
	}
	token := os.Getenv(envCLIThreadReadServiceToken)
	if token == strings.TrimSpace(os.Getenv(envExternalAuthzAPIKey)) {
		return nil, fmt.Errorf("CLI thread read requires a dedicated service credential")
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
			return nil, fmt.Errorf("CLI thread read cannot reuse a static service credential")
		}
	}
	authorizer, err := authpep.NewCLIThreadReadAuthorizer(authpep.CLIThreadReadConfig{MonolithOrigin: os.Getenv(envCLIThreadReadMonolithOrigin), Resource: os.Getenv(envCLIThreadReadResource), ServiceToken: token})
	if err != nil {
		return nil, err
	}
	return authorizer, nil
}

func runtimeCLIThreadRead(config runtimeConfig) (runtimeConfig, error) {
	authorizer, err := cliThreadReadAuthorizerFromEnv()
	if err != nil {
		return runtimeConfig{}, err
	}
	config.CLIThreadReadAuthorizer = authorizer
	return config, nil
}
