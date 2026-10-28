package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func newCLIInstallationLeaseFixture(t *testing.T) (*cliInstallationFixture, *DynamoDBAssignmentOperationStore, *terminationDynamoFixture, map[string]map[string]string) {
	t.Helper()
	dynamo, newStore, assignment := newTerminationDynamoFixture(t)
	dynamo.now = time.Now().UTC()
	store := newStore()
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: store})
	intent := f.intent(t)
	if w := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true); w.Code != http.StatusOK {
		t.Fatalf("confirm=%d", w.Code)
	}
	assignment.State = AssignmentRunning
	assignment.UpdatedAt = dynamo.now
	claim := executionClaim{DeviceID: "fixture-device", OwnerPrincipalID: "owner", RequesterPrincipalID: "other", AssignmentID: assignment.ID, AgentID: assignment.AgentID, LeaseEpoch: leaseEpoch(assignment.LeaseToken), ClaimedAt: assignment.UpdatedAt}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	projection := terminationProjectionItem(t, assignment, 4)
	active := store.agentActiveAssignmentDynamoDBItem(AssignmentOperationRecord{AgentID: assignment.AgentID, AssignmentID: assignment.ID, Assignment: assignment, RecordedAt: dynamo.now})
	dynamo.mu.Lock()
	dynamo.items[terminationItemKey(item)] = item
	dynamo.items[terminationItemKey(projection)] = projection
	dynamo.applyPolicyUpdate(store.devicePolicyAdvanceAction("fixture-device").Update)
	dynamo.mu.Unlock()
	return f, store, dynamo, active
}
