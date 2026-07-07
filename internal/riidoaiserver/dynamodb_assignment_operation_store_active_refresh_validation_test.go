package riidoaiserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDynamoDBAssignmentOperationStoreSkipsInactiveActiveRefresh(t *testing.T) {
	fixedNow := time.Date(2026, 7, 7, 6, 20, 0, 0, time.UTC)
	store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
		Now:           fixedNow,
		RequestBuffer: 1,
		Handler:       failOnUnexpectedDynamoDBRequest(t),
	})

	assignment := sampleAssignmentOperationRecord(fixedNow).Assignment
	assignment.State = AssignmentQueued
	err := store.RefreshAgentActiveAssignment(context.Background(), assignment, fixedNow)
	if err != nil {
		t.Fatalf("RefreshAgentActiveAssignment inactive state: %v", err)
	}
	assertNoCapturedDynamoDBRequest(t, requests)
}

func TestDynamoDBAssignmentOperationStoreRejectsInvalidActiveRefresh(t *testing.T) {
	fixedNow := time.Date(2026, 7, 7, 6, 21, 0, 0, time.UTC)
	base := sampleAssignmentOperationRecord(fixedNow).Assignment
	base.State = AssignmentRunning
	base.LeaseToken = "lease-1"
	cases := []struct {
		name string
		edit func(*Assignment)
		want string
	}{
		{"missing agent", func(a *Assignment) { a.AgentID = " " }, "agent_id and assignment_id"},
		{"missing assignment", func(a *Assignment) { a.ID = " " }, "agent_id and assignment_id"},
		{"missing lease", func(a *Assignment) { a.LeaseToken = " " }, "requires lease_token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, requests := newDynamoDBAssignmentOperationStoreHarness(t, dynamoDBAssignmentOperationStoreHarnessConfig{
				Now:           fixedNow,
				RequestBuffer: 1,
				Handler:       failOnUnexpectedDynamoDBRequest(t),
			})
			assignment := base
			tc.edit(&assignment)
			err := store.RefreshAgentActiveAssignment(context.Background(), assignment, fixedNow)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("RefreshAgentActiveAssignment error = %v", err)
			}
			assertNoCapturedDynamoDBRequest(t, requests)
		})
	}
}

func failOnUnexpectedDynamoDBRequest(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(http.ResponseWriter, *http.Request) {
		t.Fatal("unexpected DynamoDB request")
	}
}

func assertNoCapturedDynamoDBRequest(t *testing.T, requests <-chan capturedDynamoDBRequest) {
	t.Helper()
	select {
	case req := <-requests:
		t.Fatalf("unexpected DynamoDB target %q", req.header.Get("X-Amz-Target"))
	default:
	}
}
