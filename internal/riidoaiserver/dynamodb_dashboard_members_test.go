package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDashboardReadsCreationKeysAndRetainsLegacyRuns(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	snapshot := StoreSnapshot{Assignments: []Assignment{{ID: "legacy", TaskID: "t", AgentID: "agent", State: AssignmentCompleted}}, Events: map[string][]TaskEvent{}}
	fresh := Assignment{ID: "remote", TaskID: "t", AgentID: "agent", State: AssignmentRunning, UpdatedAt: now}
	var gets, queries int
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			KeyConditionExpression, ProjectionExpression string
			ExpressionAttributeValues, Key               map[string]map[string]string
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
			return
		}
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBQueryTarget:
			queries++
			if input.KeyConditionExpression != "pk = :pk AND begins_with(sk, :prefix)" || input.ExpressionAttributeValues[":pk"]["S"] != agentHistoryPK("agent") || input.ExpressionAttributeValues[":prefix"]["S"] != "assign:" || input.ProjectionExpression != "assignment_id" {
				t.Errorf("unbounded history read: %+v", input)
			}
			_, _ = w.Write([]byte(`{"Items":[{"assignment_id":{"S":"legacy"}},{"assignment_id":{"S":"remote"}}]}`))
		case dynamoDBGetItemTarget:
			gets++
			if input.Key["pk"]["S"] != dynamoDBAssignmentProjectionPK("remote") {
				t.Error("reread terminal projection")
			}
			if err := json.NewEncoder(w).Encode(map[string]any{"Item": sampleAssignmentProjectionDynamoDBItem(t, fresh, 9)}); err != nil {
				t.Error(err)
			}
		default:
			t.Error("unexpected write")
		}
	}))
	defer endpoint.Close()
	creds, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	db, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{TableName: "table", Region: "test", Endpoint: endpoint.URL, CredentialsProvider: creds, EnableAgentHistory: true})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	view, missing, err := db.dashboardSnapshot(ctx, "agent", snapshot, false, "")
	if err != nil || len(view.Assignments) != 2 || missing["remote"] != 9 || queries != 1 || gets != 1 {
		t.Fatalf("snapshot: %v, pending=%v calls=%d/%d", err, missing, queries, gets)
	}
	_, _, err = db.dashboardSnapshot(ctx, "agent", snapshot, true, "")
	if err != nil || queries != 1 || gets != 1 {
		t.Fatalf("idle read history, err=%v", err)
	}
}
