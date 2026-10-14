package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestDynamoDBStreamRelayCheckpointStoreNilAndClosedBoundaries(t *testing.T) {
	var nilStore *DynamoDBStreamRelayCheckpointStore
	if _, ok, err := nilStore.LoadStreamRelayCheckpoint(context.Background(), "", ""); ok || err != nil {
		t.Fatalf("nil load ok=%v err=%v", ok, err)
	}
	if err := nilStore.SaveStreamRelayCheckpoint(context.Background(), StreamRelayCheckpoint{}); err != nil {
		t.Fatalf("nil save err=%v", err)
	}
	provider := mustStaticAWSTestProvider(t, "AKID", "")
	store, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            "http://127.0.0.1:9",
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStreamRelayCheckpointStore: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	_, _, err = store.LoadStreamRelayCheckpoint(context.Background(), "stream", "shard")
	if err == nil || !strings.Contains(err.Error(), "checkpoint store closed") {
		t.Fatalf("closed load err=%v", err)
	}
	if err := store.SaveStreamRelayCheckpoint(context.Background(), StreamRelayCheckpoint{}); err == nil {
		t.Fatal("expected closed save error")
	}
}
