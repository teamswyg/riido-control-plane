package riidoaiserver

import "testing"

func TestCoverageTerminationRejectsCredentialRotationBeforeCommit(t *testing.T) {
	f, store, id := newCoveredTerminationFixture(t, "local_invocations_quiescent")
	f.beforeCommit = func(items map[string]map[string]map[string]string) {
		items["DEVICE_POLICY#dcov_fixture|REVISION"]["credential_generation"] = map[string]string{"S": "rotated"}
		items["DEVICE_POLICY#dcov_fixture|REVISION"]["revision"] = map[string]string{"N": "8"}
	}
	if _, err := store.recordExecutionTermination(t.Context(), id, coverageFixtureTermination()); err == nil {
		t.Fatal("credential rotation was omitted from receipt CAS")
	}
	if ledger := coverageFixtureLedger(t, f); ledger.Unresolved != 1 || ledger.Revision != 8 {
		t.Fatal("rejected receipt modified execution ledger")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.items["DEVICE_EXECUTION_TERMINATION#dcov_fixture|"+id.key()]) != 0 {
		t.Fatal("receipt persisted without its authority CAS")
	}
}

func TestCoverageOldReceiptCannotClearNewClaim(t *testing.T) {
	f, store, id := newCoveredTerminationFixture(t, "local_invocations_quiescent")
	old := f.items[dynamoDBAssignmentProjectionPK(id.assignmentID)+"|"+dynamoDBAssignmentProjectionSK]
	projection, err := assignmentProjectionFromDynamoDBItem(old)
	if err != nil {
		t.Fatal(err)
	}
	projection.Assignment.LeaseToken = "new-lease-fixture"
	projection.Assignment.State = AssignmentRunning
	f.items[terminationItemKey(old)] = terminationProjectionItem(t, projection.Assignment, 4)
	claim, err := executionClaimFromItem(f.items["DEVICE_EXECUTION#dcov_fixture|"+id.key()])
	if err != nil {
		t.Fatal(err)
	}
	claim.LeaseEpoch = leaseEpoch(projection.Assignment.LeaseToken)
	newClaim, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	f.items[terminationItemKey(newClaim)] = newClaim
	f.items["DEVICE_POLICY#dcov_fixture|REVISION"]["unresolved"] = map[string]string{"N": "2"}
	f.items["DEVICE_POLICY#dcov_fixture|REVISION"]["revision"] = map[string]string{"N": "8"}
	if _, err := store.recordExecutionTermination(t.Context(), id, coverageFixtureTermination()); err != nil {
		t.Fatal("original epoch was not independently closeable:", err)
	}
	if ledger := coverageFixtureLedger(t, f); ledger.Unresolved != 1 || ledger.Revision != 9 {
		t.Fatal("old epoch receipt cleared the later execution")
	}
	current, found, err := store.LoadAssignmentProjection(t.Context(), id.assignmentID)
	if err != nil || !found || current.Assignment.State != AssignmentRunning || current.Assignment.LeaseToken != "new-lease-fixture" {
		t.Fatal("old receipt changed the new running projection")
	}
}
