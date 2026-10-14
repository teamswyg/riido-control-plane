package riidoaiserver

import (
	"testing"
)

func TestDynamoDBAssignmentOperationStoreRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "ap-northeast-2", TableName: "assignments"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}
