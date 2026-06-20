package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreWritesPutItem(t *testing.T) {
	fixedNow := time.Date(2026, 5, 26, 1, 2, 3, 0, time.UTC)
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		TableName:     "riido-ai-server-assignments",
		RequestBuffer: 2,
		AccessKeyID:   "AKIDEXAMPLE",
		SessionToken:  "SESSION",
		Handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{}`))
		},
	})

	record := sampleAssignmentOperationRecord(fixedNow)
	if err := store.SaveAssignmentOperation(context.Background(), record); err != nil {
		t.Fatalf("SaveAssignmentOperation: %v", err)
	}

	first := <-requests
	second := <-requests
	operationPayload := decodeDynamoDBPutPayload(t, first)
	projectionPayload := decodeDynamoDBPutPayload(t, second)
	if operationPayload.Item["pk"]["S"] != dynamoDBAssignmentOperationPK {
		operationPayload, projectionPayload = projectionPayload, operationPayload
	}
	assertAssignmentOperationJournalPutPayload(t, operationPayload)
	assertAssignmentProjectionPutPayload(t, projectionPayload)
}
