package main

import (
	"os"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

const envCLIInstallationResource = "RIIDO_AI_SERVER_CLI_INSTALLATION_RESOURCE"

func cliInstallationServiceFromEnv() (riidoaiserver.CLIInstallationServiceConfig, error) {
	credentials, err := parseAuthzTokenCredentialsJSON(os.Getenv(envAuthzTokensJSON))
	if err != nil {
		return riidoaiserver.CLIInstallationServiceConfig{}, err
	}
	return riidoaiserver.NewCLIInstallationServiceConfig(os.Getenv(envCLIInstallationResource), credentials)
}
