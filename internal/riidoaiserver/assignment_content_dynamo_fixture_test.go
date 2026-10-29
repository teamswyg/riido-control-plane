package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type contentBundleFixture struct {
	assignmentContextFakeDynamoDB
	reject bool
	writes int
}

func (f *contentBundleFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Amz-Target") == dynamoDBPutItemTarget {
		f.putRequest(w, r)
		return
	}
	if r.Header.Get("X-Amz-Target") != dynamoDBTransactWriteTarget {
		f.assignmentContextFakeDynamoDB.ServeHTTP(w, r)
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	var request struct{ TransactItems []dynamoDBTransactWriteItem }
	if json.NewDecoder(r.Body).Decode(&request) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, item := range request.TransactItems {
		allowed := false
		if item.Put != nil {
			p := item.Put
			allowed = f.allowed(p.Item, p.ConditionExpression, p.ExpressionAttributeNames, p.ExpressionAttributeValues)
			if strings.HasPrefix(p.Item["sk"]["S"], assignmentContentPrefix) || p.Item["schema"]["S"] == "assignment_context_snapshot/v2" {
				allowed = allowed && p.ConditionExpression == "attribute_not_exists(pk)"
			}
		} else if check := item.ConditionCheck; check != nil {
			allowed = f.allowed(check.Key, check.ConditionExpression, check.ExpressionAttributeNames, check.ExpressionAttributeValues)
		}
		if !allowed {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"__type":"TransactionCanceledException"}`))
			return
		}
	}
	if f.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	for _, item := range request.TransactItems {
		if item.Put != nil {
			f.put(item.Put.Item)
		}
	}
	f.writes++
	_, _ = w.Write([]byte(`{}`))
}

func newContentBundleDynamo(t *testing.T, fixture *contentBundleFixture) *DynamoDBAssignmentOperationStore {
	t.Helper()
	endpoint := httptest.NewServer(fixture)
	t.Cleanup(endpoint.Close)
	credentials, _ := NewStaticAWSCredentialsProvider("SYNTHETIC", "SYNTHETIC", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}
