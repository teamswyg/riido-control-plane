package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func (f *snapshotDynamoDBFixture) put(t *testing.T, body []byte, w http.ResponseWriter) {
	t.Helper()
	var payload struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Errorf("decode PutItem payload: %v", err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items[payload.Item["sk"]["S"]] = payload.Item
	_, _ = w.Write([]byte(`{}`))
}

func (f *snapshotDynamoDBFixture) query(w http.ResponseWriter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]map[string]map[string]string, 0, len(f.items))
	for _, item := range f.items {
		out = append(out, item)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"Items": out})
}

func (f *snapshotDynamoDBFixture) transact(t *testing.T, body []byte, w http.ResponseWriter) {
	t.Helper()
	var payload struct {
		TransactItems []dynamoDBTransactWritePut `json:"TransactItems"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, write := range payload.TransactItems {
		if write.Put.Item["sk"]["S"] != dynamoDBAIAgentClientSnapshotSK {
			continue
		}
		expected := write.Put.ExpressionAttributeValues[":expected"]["N"]
		current := f.items[dynamoDBAIAgentClientSnapshotSK]["version"]["N"]
		if (current == "" && expected != "0") || (current != "" && current != expected) {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"__type":"TransactionCanceledException"}`))
			return
		}
	}
	for _, write := range payload.TransactItems {
		f.items[write.Put.Item["sk"]["S"]] = write.Put.Item
	}
	_, _ = w.Write([]byte(`{}`))
}
