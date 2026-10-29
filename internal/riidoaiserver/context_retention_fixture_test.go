package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type contextRetentionFixture struct {
	fake        assignmentContextFakeDynamoDB
	store       *DynamoDBAssignmentOperationStore
	now         time.Time
	beforeWrite func()
	deny        bool
	writes      int
	queryPages  []contextRetentionQueryPage
}

func newContextRetentionFixture(t *testing.T) *contextRetentionFixture {
	t.Helper()
	f := &contextRetentionFixture{now: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)}
	endpoint := httptest.NewServer(f)
	t.Cleanup(endpoint.Close)
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials, Now: func() time.Time { return f.now }, EnableAssignmentContextRetention: true})
	if err != nil {
		t.Fatal(err)
	}
	f.store = store
	t.Cleanup(func() { _ = store.Close() })
	return f
}

func (f *contextRetentionFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Amz-Target") == dynamoDBQueryTarget && len(f.queryPages) > 0 {
		page := f.queryPages[0]
		f.queryPages = f.queryPages[1:]
		_ = json.NewEncoder(w).Encode(page)
		return
	}
	if r.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget {
		f.writes++
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		var envelope struct{ TransactItems []dynamoDBTransactWriteItem }
		_ = json.Unmarshal(body, &envelope)
		if hook := f.beforeWrite; hook != nil {
			f.beforeWrite = nil
			hook()
		}
		if f.deny || f.conflicts(envelope.TransactItems) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"TransactionCanceledException"}`))
			return
		}
	}
	f.fake.ServeHTTP(w, r)
}
