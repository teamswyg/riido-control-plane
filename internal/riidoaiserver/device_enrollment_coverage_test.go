package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCoverageEnrollmentBirthIsAtomicAndServerIssued(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	native := newSnapshotDynamoDBStore(t, time.Now(), fixture.server.URL, nil)
	t.Cleanup(func() { _ = native.Close() })
	base := NewDevelopmentAIAgentClientStore()
	base.ConfigureExecutionCoverageEnrollment(true)
	store, err := OpenPersistentAIAgentClientStore(context.Background(), base, native)
	if err != nil {
		t.Fatal(err)
	}
	principal := profileColorTestPrincipal()
	response, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "coverage-new-machine", ExecutionCoverageProtocol: executionCoverageProtocol})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(response.DeviceID, executionCoveragePrefix) || response.ExecutionCoverageGeneration == "" || response.CredentialGeneration == "" {
		t.Fatal("new coverage enrollment returned an untracked credential")
	}
	fixture.mu.Lock()
	item := fixture.items["DEVICE_POLICY#"+response.DeviceID+"\x00REVISION"]
	fixture.mu.Unlock()
	coverage, err := executionCoverageFromItem(item, response.DeviceID)
	if err != nil || coverage.Generation != response.ExecutionCoverageGeneration || coverage.CredentialGeneration != response.CredentialGeneration || coverage.OwnerPrincipalID != principal.PrincipalID || coverage.State != "complete" || coverage.Revision != 1 || coverage.Unresolved != 0 {
		t.Fatal("published credential lacks matching initial durable coverage")
	}
}
