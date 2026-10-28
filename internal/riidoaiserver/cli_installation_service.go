package riidoaiserver

import "errors"

// This authority is assembled independently of user/JWT/device fallback chains.
type CLIInstallationServiceConfig struct {
	Resource   string
	authorizer *StaticTokenAuthorizer
}

func NewCLIInstallationServiceConfig(resource string, credentials []StaticTokenCredential) (CLIInstallationServiceConfig, error) {
	var selected []StaticTokenCredential
	for _, credential := range credentials {
		if credential.PrincipalID != cliInstallationServicePrincipal {
			continue
		}
		for _, scope := range credential.Scopes {
			if scope != "cli-installation:link" && scope != "cli-installation:policy-read" {
				return CLIInstallationServiceConfig{}, errors.New("CLI installation service requires exact operation scopes")
			}
		}
		selected = append(selected, credential)
	}
	if len(selected) == 0 && resource == "" {
		return CLIInstallationServiceConfig{}, nil
	}
	if len(selected) == 0 || !cliInstallationResource(resource) {
		return CLIInstallationServiceConfig{}, errors.New("CLI installation service requires exact resource and dedicated principal")
	}
	authorizer, err := NewStaticTokenAuthorizer(selected)
	if err != nil {
		return CLIInstallationServiceConfig{}, errors.New("invalid CLI installation service credentials")
	}
	return CLIInstallationServiceConfig{Resource: resource, authorizer: authorizer}, nil
}
