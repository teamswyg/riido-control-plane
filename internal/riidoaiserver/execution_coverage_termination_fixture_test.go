package riidoaiserver

import "testing"

func newCoveredTerminationFixture(t *testing.T, scope string) (*terminationDynamoFixture, *DynamoDBAssignmentOperationStore, executionTerminationIdentity) {
	t.Helper()
	f, open, a := newTerminationDynamoFixture(t)
	coverage, err := executionCoverageFromItem(coverageRevisionFixture(), "dcov_fixture")
	if err != nil {
		t.Fatal(err)
	}
	coverage.Unresolved = 1
	claim := executionClaim{DeviceID: coverage.DeviceID, OwnerPrincipalID: "owner", RequesterPrincipalID: "other", AssignmentID: a.ID, AgentID: a.AgentID, LeaseEpoch: leaseEpoch(a.LeaseToken), ClaimedAt: a.UpdatedAt, CoverageGeneration: coverage.Generation, CredentialGeneration: coverage.CredentialGeneration, RuntimeProvider: "codex", TerminationScope: scope}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := executionCoverageItem(coverage)
	if err != nil {
		t.Fatal(err)
	}
	f.items[terminationItemKey(item)] = item
	f.items[terminationItemKey(ledger)] = ledger
	id := executionTerminationIdentity{deviceID: coverage.DeviceID, ownerPrincipalID: "owner", assignmentID: a.ID, epoch: claim.LeaseEpoch, coverage: executionCoverageAuth{coverage.Generation, coverage.CredentialGeneration}}
	return f, open(), id
}

func coverageFixtureLedger(t *testing.T, f *terminationDynamoFixture) executionCoverage {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	value, err := executionCoverageFromItem(f.items["DEVICE_POLICY#dcov_fixture|REVISION"], "dcov_fixture")
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func coverageFixtureTermination() executionTerminationRequest {
	return executionTerminationRequest{SchemaVersion: "riido-execution-termination.v1", Kind: "local_invocations_quiescent", StartedInvocations: 2}
}
