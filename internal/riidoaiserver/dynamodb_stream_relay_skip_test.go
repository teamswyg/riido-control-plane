package riidoaiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	publisher := &fakeStreamRelayPublisher{}
	stats, err := RunDynamoDBStreamRelayOnce(context.Background(), DynamoDBStreamRelayConfig{
		Region: "ap-northeast-2", StreamARN: streamRelayTestARN, Endpoint: server.URL,
		CredentialsProvider: mustStaticAWSTestProvider(t, "AKID", ""),
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
