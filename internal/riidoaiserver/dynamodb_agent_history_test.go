package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestAgentHistoryIsScopedOutsideMutationActor(t *testing.T) {
	entered := make(chan struct{})
	var once sync.Once
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Target") == dynamoDBQueryTarget {
			var p struct {
				ExpressionAttributeValues map[string]map[string]string
				IndexName                 string
			}
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				t.Error(err)
			}
			if p.IndexName != "" || p.ExpressionAttributeValues[":pk"]["S"] != agentHistoryPK("agent") {
				t.Error("history used global journal or queue index")
			}
			once.Do(func() { close(entered) })
			<-r.Context().Done()
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer endpoint.Close()
	credentials, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials, EnableAgentHistory: true})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := store.LoadAgentAssignmentOperations(ctx, "agent"); done <- err }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("query never started")
	}
	writeCtx, writeCancel := context.WithTimeout(context.Background(), time.Second)
	defer writeCancel()
	if err := store.SaveAssignmentOperation(writeCtx, sampleAssignmentOperationRecord(time.Now())); err != nil {
		t.Fatalf("history blocked mutations: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled history succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("history ignored deadline")
	}
}
