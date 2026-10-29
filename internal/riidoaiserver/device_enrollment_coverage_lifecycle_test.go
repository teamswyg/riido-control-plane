package riidoaiserver

import (
	"context"
	"testing"
)

func TestCoverageEnrollmentRotationPreservesLedgerWithFlagOffAndLegacyRequest(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	store := openCoverageEnrollmentFixture(t, fixture.server.URL, true)
	principal := profileColorTestPrincipal()
	request := EnrollDeviceRequest{MachineID: "coverage-rotation", ExecutionCoverageProtocol: executionCoverageProtocol}
	first, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, request)
	if err != nil {
		t.Fatal(err)
	}
	fixture.mu.Lock()
	key := "DEVICE_POLICY#" + first.DeviceID + "\x00REVISION"
	fixture.items[key]["unresolved"] = map[string]string{"N": "3"}
	fixture.items[key]["revision"] = map[string]string{"N": "4"}
	fixture.mu.Unlock()
	reopened := openCoverageEnrollmentFixture(t, fixture.server.URL, false)
	request.ExecutionCoverageProtocol = ""
	second, err := reopened.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, request)
	if err != nil {
		t.Fatal(err)
	}
	if first.DeviceID != second.DeviceID || first.ExecutionCoverageGeneration != second.ExecutionCoverageGeneration || first.CredentialGeneration == second.CredentialGeneration {
		t.Fatal("rotation replaced immutable coverage or failed to rotate credential generation")
	}
	fixture.mu.Lock()
	policy, err := executionCoverageFromItem(fixture.items[key], second.DeviceID)
	fixture.mu.Unlock()
	if err != nil || policy.Revision != 5 || policy.Unresolved != 3 || policy.CredentialGeneration != second.CredentialGeneration {
		t.Fatal("rotation lost the durable ledger or did not advance its fence")
	}
	fresh := openCoverageEnrollmentFixture(t, fixture.server.URL, false)
	assertEnrollmentCredential(t, fresh, first, false)
	assertEnrollmentCredential(t, fresh, second, true)
	auth, err := fresh.AuthorizeDeviceCredential(context.Background(), second.DeviceID, second.DeviceSecret, AuthorizationRequest{Resource: AuthorizationResourceAgent})
	if err != nil || auth.executionCoverageGeneration != second.ExecutionCoverageGeneration || auth.credentialGeneration != second.CredentialGeneration {
		t.Fatal("verified device identity lost coverage provenance")
	}
}
