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
