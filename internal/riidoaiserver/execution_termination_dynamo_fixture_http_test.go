package riidoaiserver

import (
	"encoding/json"
	"io"
	"net/http"
)

func (f *terminationDynamoFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, data)
	switch r.Header.Get("X-Amz-Target") {
	case dynamoDBGetItemTarget:
		var req struct {
			Key            map[string]map[string]string
			ConsistentRead bool
		}
		_ = json.Unmarshal(data, &req)
		if !req.ConsistentRead {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": f.items[terminationItemKey(req.Key)]})
	case dynamoDBTransactWriteTarget:
		var req struct{ TransactItems []dynamoDBTransactWriteItem }
		_ = json.Unmarshal(data, &req)
		f.actions = req.TransactItems
		if hook := f.beforeCommit; hook != nil {
			f.beforeCommit = nil
			hook(f.items)
		}
		if f.deny || !f.conditionsHold(req.TransactItems) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"TransactionCanceledException"}`))
			return
		}
		for _, action := range req.TransactItems {
			if action.Put != nil {
				f.items[terminationItemKey(action.Put.Item)] = action.Put.Item
				f.commits++
			}
		}
		if f.loseACK {
			f.loseACK = false
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"lost acknowledgement"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
