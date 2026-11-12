package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
)

func TestDynamoDBAssignmentOperationStoreLoadsWorkspaceFailureFromTerminalOperation(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 456000000, time.UTC)
	record := sampleAssignmentOperationRecord(fixedNow)
	record.Assignment.State = AssignmentFailed
	record.OperationType = AssignmentOperationAgentEvent
	record.Events[0].State = AssignmentFailed
	record.Events[0].Type = EventAssignmentFailed
	record.Events[0].Metadata = map[string]string{assignment.ExecutionDiagnosticKey: workspacePreparationFailedDiagnostic}
	record.OperationID = assignmentOperationID(record.OperationType, record.Assignment, record.Events)
	projectionItem, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	operationItem, err := assignmentOperationDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ConsistentRead bool
			Key            map[string]map[string]string
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("decode get payload: %v", err)
			return
		}
		if !input.ConsistentRead {
			t.Error("assignment and terminal operation must be read consistently")
		}
		var item map[string]map[string]string
		switch {
		case dynamoDBStringValue(input.Key, "pk") == "ASSIGNMENT#asn-000001" &&
			dynamoDBStringValue(input.Key, "sk") == "STATE":
			item = projectionItem
		case dynamoDBStringValue(input.Key, "pk") == "ASSIGNMENT_OPERATION" &&
			dynamoDBStringValue(input.Key, "sk") == "20260526T010203.456000000Z#00000000000000000002#agent-event:asn-000001:2":
			item = operationItem
		default:
			t.Errorf("unexpected get key: %v", input.Key)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"Item": item})
	}))
	defer server.Close()

	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{
		Region:              "ap-northeast-2",
		TableName:           "assignments",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	projection, found, err := store.LoadAssignmentProjection(t.Context(), record.AssignmentID)
	if err != nil || !found || !projection.WorkspacePreparationFailed {
		t.Fatalf("terminal workspace failure not found: found=%v, err=%v", found, err)
	}
}
