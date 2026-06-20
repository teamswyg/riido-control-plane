package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

type dynamoDBRuntimeConfig struct {
	region   string
	endpoint string
	provider riidoaiserver.AWSCredentialsProvider
}

func dynamoDBConfigForFeature(feature string) (dynamoDBRuntimeConfig, error) {
	tableName := strings.TrimSpace(os.Getenv(envAIAgentClientTable))
	if tableName == "" {
		return dynamoDBRuntimeConfig{}, fmt.Errorf("%s is required when %s is enabled", envAIAgentClientTable, envAIAgentClientDev)
	}
	region := strings.TrimSpace(os.Getenv(envAWSRegion))
	if region == "" {
		return dynamoDBRuntimeConfig{}, fmt.Errorf("%s is required when %s is enabled", envAWSRegion, envAIAgentClientDev)
	}
	provider, err := awsContainerCredentialsProviderFromEnvFor(feature)
	if err != nil {
		return dynamoDBRuntimeConfig{}, err
	}
	return dynamoDBRuntimeConfig{
		region:   region,
		endpoint: strings.TrimSpace(os.Getenv(envDynamoDBEndpoint)),
		provider: provider,
	}, nil
}

func wrapEnvError(key string, err error) error {
	return fmt.Errorf("%s: %w", key, err)
}
