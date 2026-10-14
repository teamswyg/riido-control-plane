package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDynamoDBStreamRelayCheckpointStoreRejectsInvalidSavesBeforeRequest(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&calls, 1)
	}))
	defer server.Close()
	provider := mustStaticAWSTestProvider(t, "AKID", "")
	store, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{
		Region: "ap-northeast-2", TableName: "assignments",
		Endpoint: server.URL, CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStreamRelayCheckpointStore: %v", err)
	}
	defer store.Close()
	cases := []struct {
		name       string
		checkpoint StreamRelayCheckpoint
		want       string
	}{
		{"schema", StreamRelayCheckpoint{SchemaVersion: "v0"}, "unsupported"},
		{"stream", StreamRelayCheckpoint{ShardID: "shard-1", SequenceNumber: "1"}, "stream_arn"},
		{"shard", StreamRelayCheckpoint{StreamARN: "stream", SequenceNumber: "1"}, "shard_id"},
		{"sequence", StreamRelayCheckpoint{StreamARN: "stream", ShardID: "shard-1"}, "sequence_number"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.SaveStreamRelayCheckpoint(context.Background(), tc.checkpoint)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("SaveStreamRelayCheckpoint error = %v, want %q", err, tc.want)
			}
		})
	}
	if got := atomic.LoadInt32(&calls); got != 0 {
		t.Fatalf("unexpected DynamoDB requests = %d", got)
	}
}
