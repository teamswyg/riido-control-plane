package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func (f *snapshotDynamoDBFixture) membershipCommit(t *testing.T, body []byte, w http.ResponseWriter) {
	t.Helper()
	var payload struct {
		TransactItems []struct {
			Put struct{ Item map[string]map[string]string }
		}
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failMembershipCommit {
		http.Error(w, `{"__type":"TransactionCanceledException"}`, http.StatusBadRequest)
		return
	}
	for _, action := range payload.TransactItems {
		f.items[dynamoDBStringValue(action.Put.Item, "sk")] = action.Put.Item
	}
	_, _ = w.Write([]byte(`{}`))
}

func (f *snapshotDynamoDBFixture) membershipRead(t *testing.T, body []byte, w http.ResponseWriter) {
	t.Helper()
	var payload struct {
		TransactItems []struct {
			Get struct{ Key map[string]map[string]string }
		}
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	responses := make([]map[string]any, 0, len(payload.TransactItems))
	for _, action := range payload.TransactItems {
		responses = append(responses, map[string]any{"Item": f.items[dynamoDBStringValue(action.Get.Key, "sk")]})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"Responses": responses})
}
