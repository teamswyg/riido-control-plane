package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestAssignmentNotificationsDynamoDBIdleReadBudget(t *testing.T) {
	var mu sync.Mutex
	calls := map[string]int{}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		mu.Lock()
		calls[target]++
		mu.Unlock()
		switch target {
		case dynamoDBQueryTarget:
			if err := json.NewEncoder(w).Encode(map[string]any{"Items": []any{}}); err != nil {
				t.Error(err)
			}
		case dynamoDBGetItemTarget:
			if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("notification performed unexpected operation %s", target)
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	defer httpServer.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	operations, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "ap-northeast-2", TableName: "assignments", Endpoint: httpServer.URL, CredentialsProvider: credentials})
	if err != nil {
		t.Fatal(err)
	}
	defer operations.Close()
	store := NewStoreWithConfig(StoreConfig{OperationStore: operations})
	defer store.Close()
	for range 2 {
		if _, err := store.AssignmentSignalVersions(context.Background(), []string{"a", "b"}); err != nil {
			t.Fatal(err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if calls[dynamoDBQueryTarget] != 4 || calls[dynamoDBGetItemTarget] != 4 || len(calls) != 2 {
		t.Fatalf("unexpected idle DynamoDB read budget: %v", calls)
	}
}
