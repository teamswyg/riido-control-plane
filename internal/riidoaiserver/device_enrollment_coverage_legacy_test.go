package riidoaiserver

import (
	"context"
	"testing"
)

func TestCoverageEnrollmentNeverUpgradesLegacyAndRejectsMissingPolicy(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	store := openCoverageEnrollmentFixture(t, fixture.server.URL, true)
	principal := profileColorTestPrincipal()
	request := EnrollDeviceRequest{MachineID: "legacy-machine"}
	first, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, request)
	if err != nil {
		t.Fatal(err)
	}
	request.ExecutionCoverageProtocol = executionCoverageProtocol
	upgraded, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, request)
	if err != nil || upgraded.DeviceID != first.DeviceID || upgraded.ExecutionCoverageGeneration != "" {
		t.Fatal("legacy registration was promoted to covered")
	}
	request.MachineID = "new-covered-machine"
	covered, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, request)
	if err != nil {
		t.Fatal(err)
	}
	fixture.mu.Lock()
	delete(fixture.items, "DEVICE_POLICY#"+covered.DeviceID+"\x00REVISION")
	fixture.mu.Unlock()
	if response, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, request); err == nil || response.DeviceSecret != "" {
		t.Fatal("missing coverage policy was recreated during rotation")
	}
	assertEnrollmentCredential(t, store, covered, true)
}
