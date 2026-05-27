package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBStreamRelayPublishesOutboxRecords(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 3, 4, 5, 0, time.UTC)
	recordJSON := marshalOutboxRecordForStreamTest(t, TaskEvent{
		Seq:          1,
		TaskID:       "task-a",
		AssignmentID: "asn-000001",
		AgentID:      "jykim1",
		Type:         EventAssignmentQueued,
		State:        AssignmentQueued,
		Message:      "queued",
		At:           fixedNow,
	})
	requests := make(chan capturedDynamoDBRequest, 3)
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
		w.Header().Set("Content-Type", dynamoDBJSONContentType)
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBStreamDescribeTarget:
			_, _ = w.Write([]byte(`{"StreamDescription":{"Shards":[{"ShardId":"shard-1"}]}}`))
		case dynamoDBStreamGetShardIteratorTarget:
			_, _ = w.Write([]byte(`{"ShardIterator":"iterator-1"}`))
		case dynamoDBStreamGetRecordsTarget:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Records": []map[string]any{
					{
						"eventID":   "event-1",
						"eventName": "INSERT",
						"dynamodb": map[string]any{
							"SequenceNumber": "42",
							"NewImage": map[string]any{
								"event_json": map[string]string{"S": recordJSON},
							},
						},
					},
				},
			})
		default:
			t.Errorf("unexpected target %q", r.Header.Get("X-Amz-Target"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKIDEXAMPLE", "SECRET", "SESSION")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	publisher := &fakeStreamRelayPublisher{}
	checkpoints := &fakeStreamRelayCheckpointStore{
		checkpoint: StreamRelayCheckpoint{
			SchemaVersion:  StreamRelayCheckpointSchemaVersion,
			StreamARN:      "arn:aws:dynamodb:ap-northeast-2:123456789012:table/riido-ai-server-event-outbox/stream/2026-05-26T03:04:05.000",
			ShardID:        "shard-1",
			SequenceNumber: "41",
			UpdatedAt:      fixedNow.Add(-time.Minute),
		},
		ok: true,
	}
	stats, err := RunDynamoDBStreamRelayOnce(context.Background(), DynamoDBStreamRelayConfig{
		Region:              "ap-northeast-2",
		StreamARN:           "arn:aws:dynamodb:ap-northeast-2:123456789012:table/riido-ai-server-event-outbox/stream/2026-05-26T03:04:05.000",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Publisher:           publisher,
		CheckpointStore:     checkpoints,
		Now:                 func() time.Time { return fixedNow },
		ShardIteratorType:   "TRIM_HORIZON",
		Limit:               25,
	})
	if err != nil {
		t.Fatalf("RunDynamoDBStreamRelayOnce: %v", err)
	}
	if stats.ShardsDiscovered != 1 || stats.BatchesRead != 1 || stats.RecordsRead != 1 || stats.RecordsPublished != 1 || stats.RecordsSkipped != 0 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.CheckpointsLoaded != 1 || stats.CheckpointsSaved != 1 {
		t.Fatalf("checkpoint stats = %+v", stats)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("events = %+v", publisher.events)
	}
	event := publisher.events[0]
	if event.SchemaVersion != StreamRelayEventSchemaVersion || event.ShardID != "shard-1" || event.SequenceNumber != "42" {
		t.Fatalf("relay event = %+v", event)
	}
	if event.EventID != "event-1" || event.EventName != "INSERT" {
		t.Fatalf("event metadata = %+v", event)
	}
	if event.Record.Event.TaskID != "task-a" || event.Record.Event.Type != EventAssignmentQueued {
		t.Fatalf("outbox record = %+v", event.Record)
	}

	assertStreamRelayRequest(t, <-requests, dynamoDBStreamDescribeTarget)
	iteratorRequest := <-requests
	assertStreamRelayRequest(t, iteratorRequest, dynamoDBStreamGetShardIteratorTarget)
	var iteratorPayload struct {
		ShardIteratorType string `json:"ShardIteratorType"`
		SequenceNumber    string `json:"SequenceNumber"`
	}
	if err := json.Unmarshal(iteratorRequest.body, &iteratorPayload); err != nil {
		t.Fatalf("decode iterator payload: %v", err)
	}
	if iteratorPayload.ShardIteratorType != "AFTER_SEQUENCE_NUMBER" || iteratorPayload.SequenceNumber != "41" {
		t.Fatalf("iterator payload = %+v", iteratorPayload)
	}
	assertStreamRelayRequest(t, <-requests, dynamoDBStreamGetRecordsTarget)
	if len(checkpoints.saved) != 1 || checkpoints.saved[0].SequenceNumber != "42" {
		t.Fatalf("saved checkpoints = %+v", checkpoints.saved)
	}
}

func TestDynamoDBStreamRelaySkipsRecordsWithoutEventJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBStreamDescribeTarget:
			_, _ = w.Write([]byte(`{"StreamDescription":{"Shards":[{"ShardId":"shard-1"}]}}`))
		case dynamoDBStreamGetShardIteratorTarget:
			_, _ = w.Write([]byte(`{"ShardIterator":"iterator-1"}`))
		case dynamoDBStreamGetRecordsTarget:
			_, _ = w.Write([]byte(`{"Records":[{"eventID":"event-1","eventName":"MODIFY","dynamodb":{"SequenceNumber":"42","NewImage":{"other":{"S":"value"}}}}]}`))
		default:
			t.Errorf("unexpected target %q", r.Header.Get("X-Amz-Target"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	publisher := &fakeStreamRelayPublisher{}
	stats, err := RunDynamoDBStreamRelayOnce(context.Background(), DynamoDBStreamRelayConfig{
		Region:              "ap-northeast-2",
		StreamARN:           "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-05-26T03:04:05.000",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Publisher:           publisher,
	})
	if err != nil {
		t.Fatalf("RunDynamoDBStreamRelayOnce: %v", err)
	}
	if stats.RecordsRead != 1 || stats.RecordsPublished != 0 || stats.RecordsSkipped != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("events = %+v", publisher.events)
	}
}

func TestDynamoDBStreamRelayAdvancesThroughEmptyPages(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 9, 10, 11, 0, time.UTC)
	recordJSON := marshalOutboxRecordForStreamTest(t, TaskEvent{
		Seq:    1,
		TaskID: "task-after-empty",
		Type:   EventAssignmentQueued,
		At:     fixedNow,
	})
	getRecordsCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBStreamDescribeTarget:
			_, _ = w.Write([]byte(`{"StreamDescription":{"Shards":[{"ShardId":"shard-1"}]}}`))
		case dynamoDBStreamGetShardIteratorTarget:
			_, _ = w.Write([]byte(`{"ShardIterator":"iterator-1"}`))
		case dynamoDBStreamGetRecordsTarget:
			getRecordsCalls++
			if getRecordsCalls == 1 {
				_, _ = w.Write([]byte(`{"NextShardIterator":"iterator-2","Records":[]}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Records": []map[string]any{
					{
						"eventID":   "event-after-empty",
						"eventName": "INSERT",
						"dynamodb": map[string]any{
							"SequenceNumber": "84",
							"NewImage": map[string]any{
								"event_json": map[string]string{"S": recordJSON},
							},
						},
					},
				},
			})
		default:
			t.Errorf("unexpected target %q", r.Header.Get("X-Amz-Target"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	publisher := &fakeStreamRelayPublisher{}
	stats, err := RunDynamoDBStreamRelayOnce(context.Background(), DynamoDBStreamRelayConfig{
		Region:              "ap-northeast-2",
		StreamARN:           "arn:aws:dynamodb:ap-northeast-2:123456789012:table/events/stream/2026-05-26T03:04:05.000",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Publisher:           publisher,
		PollInterval:        time.Nanosecond,
	})
	if err != nil {
		t.Fatalf("RunDynamoDBStreamRelayOnce: %v", err)
	}
	if stats.BatchesRead != 2 || stats.RecordsPublished != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	if len(publisher.events) != 1 || publisher.events[0].Record.Event.TaskID != "task-after-empty" {
		t.Fatalf("events = %+v", publisher.events)
	}
}

func TestDynamoDBStreamRelayRejectsInvalidConfig(t *testing.T) {
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	publisher := &fakeStreamRelayPublisher{}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{StreamARN: "arn", CredentialsProvider: provider, Publisher: publisher}); err == nil {
		t.Fatal("expected missing region error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{Region: "ap-northeast-2", CredentialsProvider: provider, Publisher: publisher}); err == nil {
		t.Fatal("expected missing stream ARN error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{Region: "ap-northeast-2", StreamARN: "arn", Publisher: publisher}); err == nil {
		t.Fatal("expected missing credentials provider error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{Region: "ap-northeast-2", StreamARN: "arn", CredentialsProvider: provider}); err == nil {
		t.Fatal("expected missing publisher error")
	}
	if _, err := NewDynamoDBStreamRelay(DynamoDBStreamRelayConfig{Region: "ap-northeast-2", StreamARN: "arn", CredentialsProvider: provider, Publisher: publisher, Limit: 1001}); err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func marshalOutboxRecordForStreamTest(t *testing.T, event TaskEvent) string {
	t.Helper()
	record, err := json.Marshal(OutboxRecord{SchemaVersion: OutboxRecordSchemaVersion, Event: event})
	if err != nil {
		t.Fatalf("marshal outbox record: %v", err)
	}
	return string(record)
}

func assertStreamRelayRequest(t *testing.T, request capturedDynamoDBRequest, wantTarget string) {
	t.Helper()
	if request.method != http.MethodPost {
		t.Fatalf("method = %s", request.method)
	}
	if request.header.Get("Content-Type") != dynamoDBJSONContentType {
		t.Fatalf("content-type = %q", request.header.Get("Content-Type"))
	}
	if request.header.Get("X-Amz-Target") != wantTarget {
		t.Fatalf("target = %q", request.header.Get("X-Amz-Target"))
	}
	authorization := request.header.Get("Authorization")
	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260526/ap-northeast-2/dynamodb/aws4_request") {
		t.Fatalf("authorization = %q", authorization)
	}
	if !strings.Contains(authorization, "SignedHeaders=content-type;host;x-amz-date;x-amz-security-token;x-amz-target") {
		t.Fatalf("authorization signed headers = %q", authorization)
	}
}

type fakeStreamRelayPublisher struct {
	events []StreamRelayEvent
}

func (p *fakeStreamRelayPublisher) PublishStreamRelayEvent(_ context.Context, event StreamRelayEvent) error {
	p.events = append(p.events, event)
	return nil
}

type fakeStreamRelayCheckpointStore struct {
	checkpoint StreamRelayCheckpoint
	ok         bool
	saved      []StreamRelayCheckpoint
}

func (s *fakeStreamRelayCheckpointStore) LoadStreamRelayCheckpoint(_ context.Context, streamARN, shardID string) (StreamRelayCheckpoint, bool, error) {
	if s.checkpoint.StreamARN != "" && s.checkpoint.StreamARN != streamARN {
		return StreamRelayCheckpoint{}, false, nil
	}
	if s.checkpoint.ShardID != "" && s.checkpoint.ShardID != shardID {
		return StreamRelayCheckpoint{}, false, nil
	}
	return s.checkpoint, s.ok, nil
}

func (s *fakeStreamRelayCheckpointStore) SaveStreamRelayCheckpoint(_ context.Context, checkpoint StreamRelayCheckpoint) error {
	s.saved = append(s.saved, checkpoint)
	return nil
}
