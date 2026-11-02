package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDashboardRemoteCompletionPrecedesLateProviderSession(t *testing.T) {
	first := time.Date(2026, 10, 30, 14, 0, 0, 0, time.UTC)
	late := first.Add(3 * time.Hour) // next day in Seoul
	a := Assignment{ID: "remote", AgentID: "agent", TaskID: "task", State: AssignmentCompleted, UpdatedAt: late}
	var historyReads int
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query struct {
			FilterExpression, ProjectionExpression string
			ExpressionAttributeValues              map[string]map[string]string
		}
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			t.Error(err)
			return
		}
		data := map[string]any{}
		switch {
		case r.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget:
			data["Item"] = sampleAssignmentProjectionDynamoDBItem(t, a, 9)
		case query.ProjectionExpression == "assignment_id":
			data["Items"] = []map[string]any{{"assignment_id": map[string]string{"S": "remote"}}}
		default:
			historyReads++
			if query.FilterExpression == "" {
				t.Error("unbounded terminal history read")
			}
			if query.ExpressionAttributeValues[":prefix"]["S"] == "agent-event:remote:" {
				events := []TaskEvent{{AssignmentID: "remote", AgentID: "agent", TaskID: "task", State: AssignmentCompleted, Seq: 9, At: late}, {AssignmentID: "remote", AgentID: "agent", TaskID: "task", State: AssignmentCompleted, Seq: 7, At: first}}
				raw, err := json.Marshal(events)
				if err != nil {
					t.Error(err)
					return
				}
				data["Items"] = []map[string]any{{"events_json": map[string]string{"S": string(raw)}}}
			}
		}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			t.Error(err)
		}
	}))
	defer api.Close()
	creds, _ := NewStaticAWSCredentialsProvider("test", "test", "")
	db, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{TableName: "table", Region: "test", Endpoint: api.URL, CredentialsProvider: creds, EnableAgentHistory: true})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for range 2 {
		snapshot, _, err := db.dashboardSnapshot(context.Background(), "agent", StoreSnapshot{Events: map[string][]TaskEvent{}}, false, "")
		if err != nil {
			t.Fatal(err)
		}
		view := buildAgentDashboard(AgentClientRecord{}, dashboardHistoryFromSnapshot(snapshot, "agent"), map[string]AIAgentTaskContextComponent{"task": {ID: "task"}}, 1, late)
		if len(view.History) != 1 || view.History[0].EndedAt == nil || !view.History[0].EndedAt.Equal(first) {
			t.Fatal("late metadata moved the completed run")
		}
		for _, day := range view.Activity {
			if day.Completed != 0 && day.Date != "2026-10-30" {
				t.Fatal("incorrect activity day", day)
			}
		}
	}
	if historyReads != len(dashboardOperationPrefixes) {
		t.Fatal("immutable terminal timestamp was reread")
	}
}
