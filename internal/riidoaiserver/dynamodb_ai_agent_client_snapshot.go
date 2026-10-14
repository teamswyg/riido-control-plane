package riidoaiserver

import (
	"net/http"
	"time"
)

const (
	dynamoDBAIAgentClientSnapshotPK = "AI_AGENT_CLIENT#snapshot"
	dynamoDBAIAgentClientSnapshotSK = "CURRENT"
)

type DynamoDBAIAgentClientSnapshotConfig struct {
	Region              string
	TableName           string
	Endpoint            string
	HTTPClient          *http.Client
	Now                 func() time.Time
	CredentialsProvider AWSCredentialsProvider
	Metrics             *AIAgentClientPersistenceMetrics
}

type DynamoDBAIAgentClientSnapshot struct {
	commands            chan dynamoDBAIAgentClientSnapshotCommand
	done                chan struct{}
	region              string
	tableName           string
	endpoint            string
	endpointHost        string
	httpClient          *http.Client
	now                 func() time.Time
	credentialsProvider AWSCredentialsProvider
	metrics             *AIAgentClientPersistenceMetrics
	partHashes          map[string]string
}
