package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func (f *coverageLifecycleDynamo) handle(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	var req coverageLifecycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.Error("invalid Dynamo fixture JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	target := r.Header.Get("X-Amz-Target")
	f.runPolicyFenceHook(target, req)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests[target]++
	write := func(value any) { _ = json.NewEncoder(w).Encode(value) }
	reject := func() {
		w.WriteHeader(http.StatusBadRequest)
		write(map[string]string{"__type": "TransactionCanceledException"})
	}
	if req.TableName != "" && !f.sameTable(t, req.TableName) {
		reject()
		return
	}
	switch target {
	case dynamoDBGetItemTarget:
		write(map[string]any{"Item": f.items[membershipDynamoKey(req.Key)]})
	case dynamoDBSnapshotTransactGetTarget:
		f.transactGet(t, req, write, reject)
	case dynamoDBQueryTarget:
		f.query(t, req, write, reject)
	case dynamoDBPutItemTarget, dynamoDBDeleteItemTarget:
		f.putOrDelete(t, req, write, reject, target)
	case dynamoDBTransactWriteTarget:
		f.transactWrite(t, req, write, reject)
	default:
		t.Errorf("unsupported Dynamo target %s", target)
		w.WriteHeader(http.StatusBadRequest)
	}
}
