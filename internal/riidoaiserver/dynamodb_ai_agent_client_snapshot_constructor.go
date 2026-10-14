package riidoaiserver

import (
	"errors"
	"strings"
)

func NewDynamoDBAIAgentClientSnapshot(config DynamoDBAIAgentClientSnapshotConfig) (*DynamoDBAIAgentClientSnapshot, error) {
	region, tableName, err := validateDynamoDBAIAgentClientSnapshotConfig(config)
	if err != nil {
		return nil, err
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	store := &DynamoDBAIAgentClientSnapshot{
		commands:            make(chan dynamoDBAIAgentClientSnapshotCommand),
		done:                make(chan struct{}),
		region:              region,
		tableName:           tableName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		credentialsProvider: config.CredentialsProvider,
		metrics:             config.Metrics,
	}
	go store.loop()
	return store, nil
}

func validateDynamoDBAIAgentClientSnapshotConfig(config DynamoDBAIAgentClientSnapshotConfig) (string, string, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return "", "", errors.New("riidoaiserver: DynamoDB AI Agent client snapshot region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return "", "", errors.New("riidoaiserver: DynamoDB AI Agent client snapshot table name is required")
	}
	if config.CredentialsProvider == nil {
		return "", "", errors.New("riidoaiserver: DynamoDB AI Agent client snapshot credentials provider is required")
	}
	return region, tableName, nil
}
