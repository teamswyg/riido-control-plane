package riidoaiserver

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type terminationDynamoFixture struct {
	mu           sync.Mutex
	items        map[string]map[string]map[string]string
	actions      []dynamoDBTransactWriteItem
	requests     [][]byte
	beforeCommit func(map[string]map[string]map[string]string)
	deny         bool
	loseACK      bool
	commits      int
	now          time.Time
}

func newTerminationDynamoFixture(t *testing.T) (*terminationDynamoFixture, func() *DynamoDBAssignmentOperationStore, Assignment) {
	t.Helper()
	f := &terminationDynamoFixture{items: map[string]map[string]map[string]string{}, now: time.Date(2026, 10, 28, 1, 0, 0, 0, time.UTC)}
	a := sampleQueuedAssignmentOperationRecord(f.now).Assignment
	a.State = AssignmentFailed
	a.LeaseToken = "lease-sentinel"
	a.CreatedBy = "requester"
	claim := executionClaim{DeviceID: "device", OwnerPrincipalID: "owner", RequesterPrincipalID: "requester", AssignmentID: a.ID, AgentID: a.AgentID, LeaseEpoch: leaseEpoch(a.LeaseToken), ClaimedAt: a.UpdatedAt}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	f.items[terminationItemKey(item)] = item
	projection := terminationProjectionItem(t, a, 3)
	f.items[terminationItemKey(projection)] = projection
	endpoint := httptest.NewServer(f)
	t.Cleanup(endpoint.Close)
	credentials, _ := NewStaticAWSCredentialsProvider("fixture", "fixture", "")
	newStore := func() *DynamoDBAssignmentOperationStore {
		s, err := NewDynamoDBAssignmentOperationStore(DynamoDBAssignmentOperationStoreConfig{Region: "test", TableName: "test", Endpoint: endpoint.URL, CredentialsProvider: credentials, Now: func() time.Time { return f.now }})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = s.Close() })
		return s
	}
	return f, newStore, a
}

func terminationItemKey(item map[string]map[string]string) string {
	return item["pk"]["S"] + "|" + item["sk"]["S"]
}

func terminationProjectionItem(t *testing.T, a Assignment, seq int64) map[string]map[string]string {
	t.Helper()
	item := sampleAssignmentProjectionDynamoDBItem(t, a, seq)
	item["lease_epoch"] = map[string]string{"S": leaseEpoch(a.LeaseToken)}
	item["assignment_state"] = map[string]string{"S": string(a.State)}
	return item
}
