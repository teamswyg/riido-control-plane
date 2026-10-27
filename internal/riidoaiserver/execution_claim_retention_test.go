package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestExecutionClaimRetentionKeepsSixBaseActionsAndBoundedSidecar(t *testing.T) {
	f := newContextRetentionFixture(t)
	record := sampleQueuedAssignmentOperationRecord(f.now)
	record.Assignment.State = AssignmentFailed
	record.Assignment.UpdatedAt = f.now
	projection, err := assignmentProjectionDynamoDBItem(record)
	if err != nil {
		t.Fatal(err)
	}
	items := []dynamoDBTransactWriteItem{{Put: &dynamoDBTransactWritePutAction{TableName: "test", Item: projection}}}
	for i := 1; i < 6; i++ {
		items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: "test", Item: map[string]map[string]string{"pk": {"S": fmt.Sprint(i)}, "sk": {"S": "fixture"}}}})
	}
	claim := executionClaim{DeviceID: "device", OwnerPrincipalID: "owner", RequesterPrincipalID: "requester", AssignmentID: "new-assignment", AgentID: "agent", LeaseEpoch: leaseEpoch("fixture-lease"), ClaimedAt: f.now}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	items = append(items, dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: "test", Item: item, ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}})
	payload, _ := json.Marshal(struct{ TransactItems []dynamoDBTransactWriteItem }{items})
	credentials, _ := f.store.credentialsProvider.Credentials(context.Background())
	_, err = f.store.writeAssignmentWithContextRetention(context.Background(), dynamoDBRequest{endpoint: f.store.endpoint, endpointHost: f.store.endpointHost, region: f.store.region, target: dynamoDBTransactWriteTarget, payload: payload, credentials: credentials, httpClient: f.store.httpClient, now: f.store.now})
	if err != nil {
		t.Fatal(err)
	}
	if f.writes != 1 {
		t.Fatalf("writes=%d", f.writes)
	}
	contextBytes := contextSnapshotItemBytes + assignmentContextMaxCommentPages*contextCommentItemBytes + contextLifecycleItemBytes + contextQuotaItemBytes
	if contextBytes+contextTerminalBaseActions*(400<<10)+executionClaimItemBytes >= 4<<20 {
		t.Fatal("aggregate budget exceeded")
	}
}
