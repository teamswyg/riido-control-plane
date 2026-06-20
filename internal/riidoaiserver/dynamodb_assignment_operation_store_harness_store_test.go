package riidoaiserver

import (
	"testing"
	"time"
)

func newAssignmentOperationStoreHarnessStore(
	t *testing.T,
	cfg dynamoDBAssignmentOperationStoreHarnessConfig,
	endpoint string,
) *DynamoDBAssignmentOperationStore {
	t.Helper()
	provider, err := NewStaticAWSCredentialsProvider(cfg.AccessKeyID, "SECRET", cfg.SessionToken)
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	storeCfg := DynamoDBAssignmentOperationStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           cfg.TableName,
		Endpoint:            endpoint,
		CredentialsProvider: provider,
	}
	if !cfg.Now.IsZero() {
		storeCfg.Now = func() time.Time { return cfg.Now }
	}
	store, err := NewDynamoDBAssignmentOperationStore(storeCfg)
	if err != nil {
		t.Fatalf("NewDynamoDBAssignmentOperationStore: %v", err)
	}
	return store
}
