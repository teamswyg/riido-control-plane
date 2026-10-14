package riidoaiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPublishingStreamRelayServer(t *testing.T, recordJSON string, requests chan capturedDynamoDBRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		requests <- capturedDynamoDBRequest{method: r.Method, header: r.Header.Clone(), body: body}
		w.Header().Set("Content-Type", dynamoDBJSONContentType)
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBStreamDescribeTarget:
			_, _ = w.Write([]byte(`{"StreamDescription":{"Shards":[{"ShardId":"shard-1"}]}}`))
		case dynamoDBStreamGetShardIteratorTarget:
			_, _ = w.Write([]byte(`{"ShardIterator":"iterator-1"}`))
		case dynamoDBStreamGetRecordsTarget:
			writeStreamRelayRecord(t, w, "event-1", "INSERT", "42", recordJSON)
		default:
			t.Errorf("unexpected target %q", r.Header.Get("X-Amz-Target"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func writeStreamRelayRecord(t *testing.T, w http.ResponseWriter, eventID, eventName, seq, recordJSON string) {
	t.Helper()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"Records": []map[string]any{{
			"eventID":   eventID,
			"eventName": eventName,
			"dynamodb": map[string]any{
				"SequenceNumber": seq,
				"NewImage": map[string]any{
					"event_json": map[string]string{"S": recordJSON},
				},
			},
		}},
	})
}
