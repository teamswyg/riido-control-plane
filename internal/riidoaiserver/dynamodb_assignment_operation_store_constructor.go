package riidoaiserver

import (
	"errors"
	"strings"
	"time"
)

func NewDynamoDBAssignmentOperationStore(config DynamoDBAssignmentOperationStoreConfig) (*DynamoDBAssignmentOperationStore, error) {
	region := strings.TrimSpace(config.Region)
	if region == "" {
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store region is required")
	}
	tableName := strings.TrimSpace(config.TableName)
	if tableName == "" {
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store table name is required")
	}
	if config.CredentialsProvider == nil {
		return nil, errors.New("riidoaiserver: DynamoDB assignment operation store credentials provider is required")
	}
	endpoint, endpointHost, err := normalizeDynamoDBEndpoint(region, strings.TrimSpace(config.Endpoint))
	if err != nil {
		return nil, err
	}
	activeLeaseDuration := config.ActiveLeaseDuration
	if activeLeaseDuration <= 0 {
		activeLeaseDuration = time.Duration(DefaultAssignmentActiveLeaseSeconds) * time.Second
	}
	store := &DynamoDBAssignmentOperationStore{
		commands:            make(chan dynamoDBAssignmentOperationCommand),
		done:                make(chan struct{}),
		region:              region,
		tableName:           tableName,
		endpoint:            endpoint,
		endpointHost:        endpointHost,
		httpClient:          dynamoDBHTTPClient(config.HTTPClient),
		now:                 dynamoDBClock(config.Now),
		activeLeaseDuration: activeLeaseDuration,
		credentialsProvider: config.CredentialsProvider,
	}
	go store.loop()
	return store, nil
}
