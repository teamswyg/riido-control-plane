package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBRecoveryWatermarkSurvivesLegacyTerminalRepair(t *testing.T) {
	now := time.Now().UTC()
	record := sampleAssignmentOperationRecord(now)
	req := recoveryEventRequest(record.Assignment)
	record.Events = []TaskEvent{{Seq: 2, Type: EventRiidoLog, State: AssignmentLeased, At: now, Metadata: req.Metadata}}
	currentItem, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	response, _ := json.Marshal(map[string]any{"Item": currentItem})
	puts := 0
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now: now, RequestBuffer: 5,
		Handler: func(w http.ResponseWriter, r *http.Request) {
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
		},
	})
	record.Assignment.State = AssignmentCompleted
	record.Events = []TaskEvent{{Seq: 3, Type: EventAssignmentCompleted, TaskID: record.TaskID, AssignmentID: record.AssignmentID, AgentID: record.AgentID, At: now}}
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		<-requests
	}
	preserved := decodeDynamoDBPutPayload(t, <-requests)
	if !strings.Contains(preserved.ConditionExpression, ":expected_seq") || preserved.Item["recovery_watermark_json"]["S"] != currentItem["recovery_watermark_json"]["S"] {
		t.Fatal("legacy repair overwrote the durable receipt or skipped CAS")
	}
}
