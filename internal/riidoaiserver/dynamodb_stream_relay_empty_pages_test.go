package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamoDBStreamRelayAdvancesThroughEmptyPages(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 9, 10, 11, 0, time.UTC)
	recordJSON := marshalOutboxRecordForStreamTest(t, TaskEvent{
		Seq: 1, TaskID: "task-after-empty", Type: EventAssignmentQueued, At: fixedNow,
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
			writeStreamRelayRecord(t, w, "event-after-empty", "INSERT", "84", recordJSON)
		default:
			t.Errorf("unexpected target %q", r.Header.Get("X-Amz-Target"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	publisher := &fakeStreamRelayPublisher{}
	stats, err := RunDynamoDBStreamRelayOnce(context.Background(), DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", StreamARN: streamRelayTestARN, Endpoint: server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
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
