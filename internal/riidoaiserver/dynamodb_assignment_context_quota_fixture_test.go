package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// Implements atomic UpdateItem conditions over the existing fixture's items.
// All other operations use the retention fixture, including terminal fences.
type quotaDynamoFixture struct {
	*contextRetentionFixture
	t *testing.T
}

func (f quotaDynamoFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Amz-Target") != dynamoDBUpdateItemTarget {
		f.contextRetentionFixture.ServeHTTP(w, r)
		return
	}
	var req struct {
		Key                                   map[string]map[string]string
		UpdateExpression, ConditionExpression string
		ExpressionAttributeNames              map[string]string
		ExpressionAttributeValues             map[string]map[string]string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		f.t.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.UpdateExpression != "SET #count = if_not_exists(#count, :zero) + :one" || !strings.Contains(req.ConditionExpression, "#count < :limit") || !strings.Contains(req.ConditionExpression, "attribute_not_exists(#ttl)") || req.ExpressionAttributeValues[":limit"]["N"] != "100" || req.ExpressionAttributeNames["#ttl"] != assignmentContextTTLAttribute {
		f.t.Error("quota must be one bounded atomic update with terminal fence")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.fake.mu.Lock()
	defer f.fake.mu.Unlock()
	key := req.Key["pk"]["S"] + "|" + req.Key["sk"]["S"]
	item := f.fake.items[key]
	count, _ := strconv.Atoi(item["read_attempts"]["N"])
	if count >= 100 || item[assignmentContextTTLAttribute] != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException"}`))
		return
	}
	if item == nil {
		f.fake.put(req.Key)
		item = f.fake.items[key]
	}
	item["read_attempts"] = map[string]string{"N": strconv.Itoa(count + 1)}
	_, _ = w.Write([]byte(`{}`))
}

func newQuotaDynamoFixture(t *testing.T) (*contextRetentionFixture, func() *DynamoDBAssignmentOperationStore) {
	t.Helper()
	f := newContextRetentionFixture(t)
	endpoint := httptest.NewServer(quotaDynamoFixture{f, t})
	t.Cleanup(endpoint.Close)
	newStore := func() *DynamoDBAssignmentOperationStore {
		store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: f.store.credentialsProvider, EnableAssignmentContextRetention: true})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = store.Close() })
		return store
	}
	f.store = newStore()
	return f, newStore
}
