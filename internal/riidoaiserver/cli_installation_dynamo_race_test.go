package riidoaiserver

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestCLIInstallationDynamoPolicyFencesClaimBeforeCommit(t *testing.T) {
	dynamo, newStore, assignment := newTerminationDynamoFixture(t)
	dynamo.now = time.Now().UTC()
	store := newStore()
	f := newCLIInstallationFixture(t, StoreConfig{OperationStore: store})
	intent := f.intent(t)
	w := f.post(t, "/v1/daemon/cli-installation-links/"+intent.IntentID+"/confirm", map[string]string{"schema_version": "riido-cli-installation-link-confirm.v1", "signature": f.signature(t, intent)}, true)
	if w.Code != http.StatusOK {
		t.Fatalf("confirm=%d", w.Code)
	}
	assignment.State = AssignmentRunning
	claim := executionClaim{DeviceID: "fixture-device", OwnerPrincipalID: "owner", RequesterPrincipalID: "requester", AssignmentID: assignment.ID, AgentID: assignment.AgentID, LeaseEpoch: leaseEpoch(assignment.LeaseToken), ClaimedAt: assignment.UpdatedAt}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	projection := terminationProjectionItem(t, assignment, 4)
	dynamo.mu.Lock()
	dynamo.beforeCommit = func(items map[string]map[string]map[string]string) {
		// A concurrent claim transaction commits after the observation query,
		// immediately before this request's final revision condition.
		items[terminationItemKey(item)] = item
		items[terminationItemKey(projection)] = projection
		dynamo.applyPolicyUpdate(store.devicePolicyAdvanceAction("fixture-device").Update)
	}
	dynamo.mu.Unlock()
	policy, err := store.readCLIInstallationPolicy(context.Background(), f.request.cliInstallationIdentity, f.devices)
	if err != nil || policy.Decision != "blocked" || policy.ExecutionRevision != "2" {
		t.Fatalf("stale policy escaped final fence: %+v %v", policy, err)
	}
}
