package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreTreatsProjectionConflictAsIdempotent(t *testing.T) {
	fixedNow := time.Date(2026, 7, 7, 5, 40, 0, 0, time.UTC)
	putCalls := 0
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 3,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Amz-Target") != dynamoDBPutItemTarget {
				_, _ = w.Write([]byte(`{}`))
				return
			}
			putCalls++
			if putCalls == 2 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#ConditionalCheckFailedException","message":"projection already current"}`))
				return
			}
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := store.SaveAssignmentOperation(context.Background(), sampleTerminalAssignmentOperationRecord(fixedNow))
	if err != nil {
		t.Fatalf("SaveAssignmentOperation terminal projection conflict: %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBDeleteItemTarget)
}

func TestDynamoDBAssignmentOperationStoreReportsProjectionWriteFailure(t *testing.T) {
	fixedNow := time.Date(2026, 7, 7, 5, 41, 0, 0, time.UTC)
	putCalls := 0
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 2,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			putCalls++
			if putCalls == 2 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"__type":"InternalServerError","Message":"projection unavailable"}`))
				return
			}
			_, _ = w.Write([]byte(`{}`))
		},
	})

	err := store.SaveAssignmentOperation(context.Background(), sampleQueuedAssignmentOperationRecord(fixedNow))
	if err == nil || !strings.Contains(err.Error(), "projection unavailable") {
		t.Fatalf("SaveAssignmentOperation projection error = %v", err)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBPutItemTarget)
}
