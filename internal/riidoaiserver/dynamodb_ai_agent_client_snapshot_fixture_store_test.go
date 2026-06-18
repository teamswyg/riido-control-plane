package riidoaiserver

import (
	"testing"
	"time"
)

func newSnapshotDynamoDBStore(t *testing.T, now time.Time, endpoint string, metrics *AIAgentClientPersistenceMetrics) *DynamoDBAIAgentClientSnapshot {
	t.Helper()
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		Endpoint:            endpoint,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return now },
		Metrics:             metrics,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	return store
}
