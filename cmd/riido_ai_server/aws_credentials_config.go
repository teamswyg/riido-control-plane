package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func awsContainerCredentialsProviderFromEnv() (riidoaiserver.AWSCredentialsProvider, error) {
	return awsContainerCredentialsProviderFromEnvFor(envAIAgentClientDev)
}

func awsContainerCredentialsProviderFromEnvFor(feature string) (riidoaiserver.AWSCredentialsProvider, error) {
	endpoint, err := awsContainerCredentialsEndpointFromEnv(feature)
	if err != nil {
		return nil, err
	}
	provider, err := riidoaiserver.NewECSContainerCredentialsProvider(riidoaiserver.ECSContainerCredentialsProviderConfig{
		Endpoint:           endpoint,
		AuthorizationToken: strings.TrimSpace(os.Getenv(envAWSContainerAuthorizationToken)),
	})
	if err != nil {
		return nil, wrapEnvError(envAWSContainerCredentialsFullURI, err)
	}
	return provider, nil
}

func awsContainerCredentialsEndpointFromEnv(feature string) (string, error) {
	endpoint := strings.TrimSpace(os.Getenv(envAWSContainerCredentialsFullURI))
	if endpoint != "" {
		return endpoint, nil
	}
	relativeURI := strings.TrimSpace(os.Getenv(envAWSContainerCredentialsRelativeURI))
	if relativeURI != "" {
		if !strings.HasPrefix(relativeURI, "/") {
			return "", fmt.Errorf("%s must start with /", envAWSContainerCredentialsRelativeURI)
		}
		return awsECSCredentialsBaseURL + relativeURI, nil
	}
	return "", fmt.Errorf("%s or %s is required when %s is configured", envAWSContainerCredentialsFullURI, envAWSContainerCredentialsRelativeURI, feature)
}
