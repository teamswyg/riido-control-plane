package riidoaiserver

import (
	"context"
	"testing"
)

func TestCoverageEnrollmentRequiresAtomicBackendAndNewBirthFlag(t *testing.T) {
	ctx := context.Background()
	principal := profileColorTestPrincipal()
	base := NewDevelopmentAIAgentClientStore()
	base.ConfigureExecutionCoverageEnrollment(true)
	request := EnrollDeviceRequest{MachineID: "memory-machine", ExecutionCoverageProtocol: executionCoverageProtocol}
	if response, err := base.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, request); err == nil || response.DeviceSecret != "" {
		t.Fatal("memory-only store issued covered credential")
	}
	fixture := newMembershipSharedDynamoFixture(t)
	disabled := openCoverageEnrollmentFixture(t, fixture.server.URL, false)
	if response, err := disabled.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, request); err == nil || response.DeviceSecret != "" {
		t.Fatal("disabled birth issued credential")
	}
	request.ExecutionCoverageProtocol = "riido-execution-coverage.v2"
	if _, err := disabled.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, request); err == nil {
		t.Fatal("unsupported coverage protocol accepted")
	}
}
