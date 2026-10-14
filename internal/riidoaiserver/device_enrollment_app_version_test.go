package riidoaiserver

import (
	"context"
	"testing"
)

func TestDeviceEnrollmentPreservesDesktopAppVersion(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "owner-a", WorkspaceID: "workspace-a"}
	_, err := store.EnrollDeviceCredential(context.Background(), principal, "workspace-a", EnrollDeviceRequest{
		MachineID: "machine-a", AppVersion: " 0.0.15 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := store.ListAIAgentDevices(context.Background(), principal)
	if err != nil || len(response.Devices) != 1 {
		t.Fatalf("ListAIAgentDevices = %+v, %v", response, err)
	}
	if got := response.Devices[0].DesktopAppVersion; got != "0.0.15" {
		t.Fatalf("desktop_app_version = %q", got)
	}
}
