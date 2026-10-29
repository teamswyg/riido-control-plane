package riidoaiserver

import "testing"

func TestCoverageClaimWithoutTrustedAdmissionFailsClosed(t *testing.T) {
	s := &DynamoDBAssignmentOperationStore{tableName: "fixture"}
	if _, err := s.appendExecutionAdmissionActions(nil, nil); err == nil {
		t.Fatal("claim without device provenance or trusted legacy binding bypassed the CAS fence")
	}
}

func TestCoverageAssignmentFinalFenceIncludesProjection(t *testing.T) {
	dynamo, newStore, assignment := newTerminationDynamoFixture(t)
	s := newStore()
	coverage, err := executionCoverageFromItem(coverageRevisionFixture(), "dcov_fixture")
	if err != nil {
		t.Fatal(err)
	}
	assignment.State = AssignmentRunning
	claim := executionClaim{DeviceID: coverage.DeviceID, OwnerPrincipalID: "owner", RequesterPrincipalID: "other", AssignmentID: assignment.ID, AgentID: assignment.AgentID, LeaseEpoch: leaseEpoch(assignment.LeaseToken), ClaimedAt: assignment.UpdatedAt, CoverageGeneration: coverage.Generation, CredentialGeneration: coverage.CredentialGeneration, RuntimeProvider: "codex", TerminationScope: coverageLocalTermination}
	item, err := executionClaimItem(claim)
	if err != nil {
		t.Fatal(err)
	}
	projection := terminationProjectionItem(t, assignment, 3)
	ledger := coverageRevisionFixture()
	dynamo.mu.Lock()
	dynamo.items[terminationItemKey(item)] = item
	dynamo.items[terminationItemKey(projection)] = projection
	dynamo.items[terminationItemKey(ledger)] = ledger
	dynamo.beforeCommit = func(items map[string]map[string]map[string]string) {
		terminal := assignment
		terminal.State = AssignmentCompleted
		items[terminationItemKey(projection)] = terminationProjectionItem(t, terminal, 4)
	}
	dynamo.mu.Unlock()
	if err := s.authorizeCoveredAssignment(t.Context(), coverage, assignment); err == nil {
		t.Fatal("stale active projection authorized dispatch after terminal commit")
	}
}
