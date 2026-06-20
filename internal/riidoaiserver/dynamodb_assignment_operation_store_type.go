package riidoaiserver

import (
	"net/http"
	"time"
)

type DynamoDBAssignmentOperationStoreConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	ActiveLeaseDuration time.Duration
	CredentialsProvider AWSCredentialsProvider
}

type DynamoDBAssignmentOperationStore struct {
	commands            chan dynamoDBAssignmentOperationCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	activeLeaseDuration time.Duration
	credentialsProvider AWSCredentialsProvider
}
