package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBStreamRelayCheckpointStoreSavesAndLoads(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 4, 5, 6, 0, time.UTC)
	streamARN := "arn:aws:dynamodb:ap-northeast-2:123456789012:table/riido-ai-server-event-outbox/stream/2026-05-26T04:05:06.000"
	requests := make(chan capturedDynamoDBRequest, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{
			method: r.Method,
			header: r.Header.Clone(),
			body:   body,
		}
		if r.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Item": map[string]any{
					"schema_version":  map[string]string{"S": StreamRelayCheckpointSchemaVersion},
					"stream_arn":      map[string]string{"S": streamARN},
					"shard_id":        map[string]string{"S": "shard-1"},
					"sequence_number": map[string]string{"S": "42"},
					"updated_at":      map[string]string{"S": fixedNow.Format(time.RFC3339Nano)},
				},
			})
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-server-assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStreamRelayCheckpointStore: %v", err)
	}
	defer store.Close()

	err = store.SaveStreamRelayCheckpoint(context.Background(), StreamRelayCheckpoint{
		StreamARN:      streamARN,
		ShardID:        "shard-1",
		SequenceNumber: "42",
	})
	if err != nil {
		t.Fatalf("SaveStreamRelayCheckpoint: %v", err)
	}
	checkpoint, ok, err := store.LoadStreamRelayCheckpoint(context.Background(), streamARN, "shard-1")
	if err != nil {
		t.Fatalf("LoadStreamRelayCheckpoint: %v", err)
	}
	if !ok {
		t.Fatal("expected checkpoint")
	}
	if checkpoint.SchemaVersion != StreamRelayCheckpointSchemaVersion || checkpoint.SequenceNumber != "42" || !checkpoint.UpdatedAt.Equal(fixedNow) {
		t.Fatalf("checkpoint = %+v", checkpoint)
	}

	putRequest := <-requests
	assertDynamoDBTarget(t, putRequest, dynamoDBPutItemTarget)
	var putPayload struct {
		TableName string                       `json:"TableName"`
		Item      map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(putRequest.body, &putPayload); err != nil {
		t.Fatalf("decode PutItem payload: %v", err)
	}
	if putPayload.TableName != "riido-ai-server-assignments" {
		t.Fatalf("put table = %q", putPayload.TableName)
	}
	assertDynamoDBString(t, putPayload.Item, "pk", streamRelayCheckpointPK(streamARN))
	assertDynamoDBString(t, putPayload.Item, "sk", streamRelayCheckpointSK("shard-1"))
	assertDynamoDBString(t, putPayload.Item, "schema_version", StreamRelayCheckpointSchemaVersion)
	assertDynamoDBString(t, putPayload.Item, "sequence_number", "42")
	assertDynamoDBString(t, putPayload.Item, "updated_at", fixedNow.Format(time.RFC3339Nano))

	getRequest := <-requests
	assertDynamoDBTarget(t, getRequest, dynamoDBGetItemTarget)
	var getPayload struct {
		TableName      string                       `json:"TableName"`
		ConsistentRead bool                         `json:"ConsistentRead"`
		Key            map[string]map[string]string `json:"Key"`
	}
	if err := json.Unmarshal(getRequest.body, &getPayload); err != nil {
		t.Fatalf("decode GetItem payload: %v", err)
	}
	if getPayload.TableName != "riido-ai-server-assignments" || !getPayload.ConsistentRead {
		t.Fatalf("get payload = %+v", getPayload)
	}
	assertDynamoDBString(t, getPayload.Key, "pk", streamRelayCheckpointPK(streamARN))
	assertDynamoDBString(t, getPayload.Key, "sk", streamRelayCheckpointSK("shard-1"))
}

func TestDynamoDBStreamRelayCheckpointStoreLoadsMissingCheckpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewDynamoDBStreamRelayCheckpointStore: %v", err)
	}
	defer store.Close()

	_, ok, err := store.LoadStreamRelayCheckpoint(context.Background(), "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-05-26T04:05:06.000", "shard-1")
	if err != nil {
		t.Fatalf("LoadStreamRelayCheckpoint: %v", err)
	}
	if ok {
		t.Fatal("expected missing checkpoint")
	}
}

func TestDynamoDBStreamRelayCheckpointStoreRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	if _, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{Region: "ap-northeast-2", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing table error")
	}
	if _, err := NewDynamoDBStreamRelayCheckpointStore(DynamoDBStreamRelayCheckpointStoreConfig{Region: "ap-northeast-2", TableName: "assignments"}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
}
