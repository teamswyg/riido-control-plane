package riidoaiserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type snapshotDynamoDBFixture struct {
	requests chan capturedDynamoDBRequest
	items    map[string]map[string]map[string]string
	server   *httptest.Server
	store    *DynamoDBAIAgentClientSnapshot
	metrics  *AIAgentClientPersistenceMetrics
}

func newSnapshotDynamoDBFixture(t *testing.T, now time.Time, items map[string]map[string]map[string]string, metrics *AIAgentClientPersistenceMetrics) *snapshotDynamoDBFixture {
	t.Helper()
	if items == nil {
		items = map[string]map[string]map[string]string{}
	}
	fixture := &snapshotDynamoDBFixture{
		requests: make(chan capturedDynamoDBRequest, 32),
		items:    items,
		metrics:  metrics,
	}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixture.handle(t, w, r)
	}))
	fixture.store = newSnapshotDynamoDBStore(t, now, fixture.server.URL, metrics)
	return fixture
}

func (f *snapshotDynamoDBFixture) handle(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("read request body: %v", err)
	}
	f.requests <- capturedDynamoDBRequest{method: r.Method, header: r.Header.Clone(), body: body}
	switch r.Header.Get("X-Amz-Target") {
	case dynamoDBPutItemTarget:
		f.put(t, body, w)
	case dynamoDBQueryTarget:
		f.query(w)
	default:
		t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
		_, _ = w.Write([]byte(`{}`))
	}
}

func (f *snapshotDynamoDBFixture) close() {
	_ = f.store.Close()
	f.server.Close()
}
