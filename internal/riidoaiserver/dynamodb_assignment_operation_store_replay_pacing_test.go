package riidoaiserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoadDynamoDBAssignmentOperationPageRetriesWithoutRestartingReplay(t *testing.T) {
	originalDelays := storeOpenRetryDelays
	storeOpenRetryDelays = []time.Duration{0}
	t.Cleanup(func() { storeOpenRetryDelays = originalDelays })

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		if calls == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"ThrottlingException","message":"retry"}`))
			return
		}
		_, _ = w.Write([]byte(`{"Items":[]}`))
	}))
	defer server.Close()

	request := dynamoDBRequest{
		endpoint:     server.URL,
		endpointHost: strings.TrimPrefix(server.URL, "http://"),
		region:       "ap-northeast-2",
		target:       dynamoDBQueryTarget,
		payload:      []byte(`{}`),
		credentials:  AWSCredentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
		httpClient:   server.Client(),
		now:          func() time.Time { return time.Date(2026, 7, 15, 4, 0, 0, 0, time.UTC) },
	}

	body, err := loadDynamoDBAssignmentOperationPage(context.Background(), request)
	if err != nil {
		t.Fatalf("loadDynamoDBAssignmentOperationPage: %v", err)
	}
	if calls != 2 || string(body) != `{"Items":[]}` {
		t.Fatalf("calls/body = %d/%q, want 2/success response", calls, body)
	}
}

func TestPaceDynamoDBAssignmentOperationReplayHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := paceDynamoDBAssignmentOperationReplay(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("paceDynamoDBAssignmentOperationReplay error = %v, want context canceled", err)
	}
}
