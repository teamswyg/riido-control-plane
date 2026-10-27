package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestExecutionClaimDynamoContentionPublishesNeitherClaimNorStart(t *testing.T) {
	f := newContextRetentionFixture(t)
	queued := sampleQueuedAssignmentOperationRecord(f.now).Assignment
	item := sampleAssignmentProjectionDynamoDBItem(t, queued, 1)
	f.fake.put(item)
	f.queryPages = []contextRetentionQueryPage{{Items: []map[string]map[string]string{item}}}
	f.deny = true
	actor := NewStoreWithConfig(StoreConfig{OperationStore: f.store, Now: func() time.Time { return f.now }})
	defer actor.Close()
	ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{"device", "owner"})
	poll, err := actor.PollAgent(ctx, queued.AgentID, PollRequest{})
	if err != nil || poll.Action != PollNone || poll.Assignment != nil {
		t.Fatal("contended claim dispatched")
	}
	claims, err := f.store.loadDeviceExecutionClaims(context.Background(), "device")
	if err != nil || len(claims) != 0 {
		t.Fatal("contended claim left orphan security row")
	}
	projection, found, err := f.store.LoadAssignmentProjection(context.Background(), queued.ID)
	if err != nil || !found || projection.Assignment.State != AssignmentQueued {
		t.Fatal("failed claim changed projection")
	}
}

func TestExecutionClaimStaleRepairAndNewClaimShareTransaction(t *testing.T) {
	fixture, store, requests := newStaleBlockedClaimFixture(t)
	ctx := context.WithValue(context.Background(), authenticatedClaimDeviceKey{}, authenticatedClaimDevice{"device", "owner"})
	claim, err := store.ClaimNextAssignment(ctx, "jykim-new", fixture.Now)
	if err != nil || !claim.Claimed {
		t.Fatalf("claim failed: %v", err)
	}
	for range 3 {
		<-requests
	}
	var payload struct{ TransactItems []dynamoDBTransactWriteItem }
	if err := json.Unmarshal((<-requests).body, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.TransactItems) != 8 {
		t.Fatalf("atomic repair count=%d", len(payload.TransactItems))
	}
	if !isDevicePolicyAdvance(payload.TransactItems[7]) {
		t.Fatal("repair revision is not in the same CAS")
	}
	if claim.Operations[0].ExecutionClaim != nil || claim.Operation.ExecutionClaim == nil {
		t.Fatal("repair manufactured or lost claim provenance")
	}
	if payload.TransactItems[6].Put == nil || dynamoDBStringValue(payload.TransactItems[6].Put.Item, "pk") != "DEVICE_EXECUTION#device" {
		t.Fatal("sidecar not in repair transaction")
	}
}
