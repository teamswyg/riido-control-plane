package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBEventACKSurvivesLegacyProjectionRepair(t *testing.T) {
	now := time.Now().UTC()
	record := sampleAssignmentOperationRecord(now)
	record.eventACK = &assignmentEventACKReceipt{KeyHash: recoveryHash("key"), PayloadHash: recoveryHash("payload"), Seq: 2, At: now, State: AssignmentReady, Type: EventAssignmentReady, AssignmentUpdatedAt: now}
	record.Events = []TaskEvent{{Seq: 2, Type: EventAssignmentReady, State: AssignmentReady, At: now}}
	current, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	response, _ := json.Marshal(map[string]any{"Item": current})
	puts := 0
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{Now: now, RequestBuffer: 5, Handler: func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Target") == dynamoDBGetItemTarget {
			_, _ = w.Write(response)
			return
		}
		if r.Header.Get("X-Amz-Target") == dynamoDBPutItemTarget {
			puts++
			if puts == 2 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"__type":"ConditionalCheckFailedException"}`))
				return
			}
		}
		_, _ = w.Write([]byte(`{}`))
	}})
	record.eventACK = nil
	record.Assignment.State = AssignmentCompleted
	record.Events = []TaskEvent{{Seq: 3, Type: EventAssignmentCompleted, TaskID: record.TaskID, AssignmentID: record.AssignmentID, AgentID: record.AgentID, At: now}}
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		<-requests
	}
	preserved := decodeDynamoDBPutPayload(t, <-requests)
	if !strings.Contains(preserved.ConditionExpression, ":expected_seq") || preserved.Item["event_ack_receipts_json"]["S"] != current["event_ack_receipts_json"]["S"] {
		t.Fatal("legacy projection repair overwrote receipt without CAS")
	}
}
