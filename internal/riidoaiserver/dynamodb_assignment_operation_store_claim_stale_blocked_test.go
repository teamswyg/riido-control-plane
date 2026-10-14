package riidoaiserver

import (
	"context"
	"testing"
)

func TestDynamoDBAssignmentOperationStoreClaimRepairsStaleBlockedAssignment(t *testing.T) {
	fixture, store, requests := newStaleBlockedClaimFixture(t)

	claim, err := store.ClaimNextAssignment(context.Background(), "jykim-new", fixture.Now)
	if err != nil {
		t.Fatalf("ClaimNextAssignment repaired: %v", err)
	}
	if !claim.Claimed || claim.Assignment.ID != fixture.Current.ID || claim.Assignment.State != AssignmentLeased || claim.Assignment.BlockedByAssignmentID != "" {
		t.Fatalf("claim = %+v", claim)
	}
	if len(claim.Operations) != 2 || claim.Operations[0].AssignmentID != fixture.Blocker.ID || claim.Operations[1].AssignmentID != fixture.Current.ID {
		t.Fatalf("claim operations = %+v", claim.Operations)
	}
	assertDynamoDBTarget(t, <-requests, dynamoDBQueryTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
	assertDynamoDBTarget(t, <-requests, dynamoDBGetItemTarget)
	transactRequest := <-requests
	assertDynamoDBTarget(t, transactRequest, dynamoDBTransactWriteTarget)
	payload := decodeDynamoDBRepairTransactWritePayload(t, transactRequest)
	assertStaleBlockedRepairPayload(t, fixture, payload)
}
