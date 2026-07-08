package riidoaiserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDynamoDBStreamRelayCheckpointStoreRejectsInvalidLoadResponses(t *testing.T) {
	streamARN := "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-05-26T04:05:06.000"
	cases := []struct {
		name string
		body string
		want string
	}{
		{"malformed", `{`, "decode DynamoDB"},
		{"schema", checkpointItemJSON(streamARN, "shard-1", "42", "", "old"), "unsupported"},
		{"sequence", checkpointItemJSON(streamARN, "shard-1", "", "", StreamRelayCheckpointSchemaVersion), "sequence_number"},
		{"updated", checkpointItemJSON(streamARN, "shard-1", "42", "bad", StreamRelayCheckpointSchemaVersion), "updated_at"},
		{"stream", checkpointItemJSON("other", "shard-1", "42", "", StreamRelayCheckpointSchemaVersion), "stream_arn mismatch"},
		{"shard", checkpointItemJSON(streamARN, "other", "42", "", StreamRelayCheckpointSchemaVersion), "shard_id mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			store := newCheckpointLoadTestStore(t, server.URL)
			defer store.Close()
			_, ok, err := store.LoadStreamRelayCheckpoint(context.Background(), streamARN, "shard-1")
			if ok || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("LoadStreamRelayCheckpoint ok=%v err=%v, want %q", ok, err, tc.want)
			}
		})
	}
}

func newCheckpointLoadTestStore(t *testing.T, endpoint string) *DynamoDBStreamRelayCheckpointStore {
	t.Helper()
	store, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            endpoint,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStreamRelayCheckpointStore: %v", err)
	}
	return store
}

func checkpointItemJSON(streamARN, shardID, sequence, updatedAt, schema string) string {
	return fmt.Sprintf(`{"Item":{
		"schema_version":{"S":%q},
		"stream_arn":{"S":%q},
		"shard_id":{"S":%q},
		"sequence_number":{"S":%q},
		"updated_at":{"S":%q}}}`,
		schema, streamARN, shardID, sequence, updatedAt,
	)
}
