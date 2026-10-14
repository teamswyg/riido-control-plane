package riidoaiserver

import (
	"context"
	"testing"
)

func TestDynamoDBStoreSnapshotRejectsBoundaryConfig(t *testing.T) {
	provider := mustStaticAWSTestProvider(t, "AKID", "")
	if _, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{
		TableName:           "assignments",
		CredentialsProvider: provider,
	}); err == nil {
		t.Fatal("expected missing region error")
	}
	_, err := NewDynamoDBStoreSnapshot(DynamoDBStoreSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            "ftp://dynamodb.local",
		CredentialsProvider: provider,
	})
	if err == nil {
		t.Fatal("expected invalid endpoint error")
	}
}

func TestDynamoDBStoreSnapshotLoopRejectsNilSaveCommand(t *testing.T) {
	store := newDynamoDBStoreSnapshotForBoundary(t, `{}`)
	defer store.Close()
	reply := make(chan error, 1)
	store.commands <- dynamoDBStoreSnapshotCommand{
		ctx:     context.Background(),
		errDone: reply,
	}
	if err := <-reply; err == nil || err.Error() != "riidoaiserver: nil DynamoDB store snapshot" {
		t.Fatalf("nil save command error = %v", err)
	}
}
