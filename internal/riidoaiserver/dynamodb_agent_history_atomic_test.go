package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAgentHistorySharesCreationAndStateCommit(t *testing.T) {
	transactions := 0
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Target") == dynamoDBTransactWriteTarget {
			var p struct{ TransactItems []dynamoDBTransactWriteItem }
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				t.Error(err)
			}
			want := 5
			if transactions == 1 {
				want = 3
			}
			if len(p.TransactItems) != want {
				t.Errorf("actions=%d want=%d", len(p.TransactItems), want)
			}
			found := false
			for _, action := range p.TransactItems {
				if action.Put != nil && action.Put.Item["pk"]["S"] == agentHistoryPK("agent-a") {
					found = true
					if action.Put.Item["agent_id"]["S"] != "" {
						t.Error("history polluted sparse queue GSI")
					}
				}
			}
			if !found {
				t.Error("history was outside commit")
			}
			transactions++
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
	record := sampleAssignmentOperationRecord(time.Now())
	record.AgentID = "agent-a"
	record.Assignment.AgentID = "agent-a"
	if err := store.saveTaskAgentCreation(context.Background(), record, credentials.credentials); err != nil {
		t.Fatal(err)
	}
	record.ExpectedProjection = &AssignmentProjection{Assignment: record.Assignment, LastEventSeq: 1}
	if err := store.saveConditionalAssignment(context.Background(), record, credentials.credentials); err != nil {
		t.Fatal(err)
	}
	if transactions != 2 {
		t.Fatal("missing commits")
	}
}
