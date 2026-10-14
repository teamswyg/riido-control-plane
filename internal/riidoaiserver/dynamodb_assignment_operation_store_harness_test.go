package riidoaiserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type dynamoDBAssignmentOperationStoreHarnessConfig struct {
	Now           time.Time
	TableName     string
	RequestBuffer int
	AccessKeyID   string
	SessionToken  string
	Handler       http.HandlerFunc
}

func newDynamoDBAssignmentOperationStoreHarness(
	t *testing.T,
	cfg dynamoDBAssignmentOperationStoreHarnessConfig,
) (*DynamoDBAssignmentOperationStore, <-chan capturedDynamoDBRequest) {
	t.Helper()
	cfg = normalizeAssignmentOperationStoreHarnessConfig(cfg)
	requests := make(chan capturedDynamoDBRequest, cfg.RequestBuffer)
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
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		cfg.Handler(w, r)
	}))
	store := newAssignmentOperationStoreHarnessStore(t, cfg, server.URL)
	t.Cleanup(func() {
		store.Close()
		server.Close()
	})
	return store, requests
}
