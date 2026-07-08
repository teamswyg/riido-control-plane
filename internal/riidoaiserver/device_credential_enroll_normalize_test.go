package riidoaiserver

import "testing"

func TestNormalizeDeviceEnrollmentTrimsAndCompletesDefaults(t *testing.T) {
	principal := AuthorizationResult{
		PrincipalID: " user-1 ",
		WorkspaceID: " workspace-principal ",
	}
	got, err := normalizeDeviceEnrollment(principal, "  ", EnrollDeviceRequest{
		DisplayName: "   ",
		MachineID:   " machine-1 ",
		DeviceID:    " device-1 ",
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got.principalID != "user-1" || got.workspaceID != "workspace-principal" {
		t.Fatalf("identity = %+v", got)
	}
	if got.displayName != "Riido Desktop" {
		t.Fatalf("displayName = %q", got.displayName)
	}
	if got.machineID != "machine-1" || got.deviceID != "device-1" {
		t.Fatalf("device identifiers = %+v", got)
	}
}

func TestNormalizeDeviceEnrollmentPrefersExplicitWorkspaceAndName(t *testing.T) {
	got, err := normalizeDeviceEnrollment(
		AuthorizationResult{PrincipalID: "user-1", WorkspaceID: "workspace-principal"},
		" workspace-request ",
		EnrollDeviceRequest{DisplayName: "  Teddy Mac  "},
	)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got.workspaceID != "workspace-request" {
		t.Fatalf("workspaceID = %q", got.workspaceID)
	}
	if got.displayName != "Teddy Mac" {
		t.Fatalf("displayName = %q", got.displayName)
	}
}

func TestNormalizeDeviceEnrollmentRejectsMissingPrincipal(t *testing.T) {
	_, err := normalizeDeviceEnrollment(
		AuthorizationResult{WorkspaceID: "workspace-principal"},
		"workspace-request",
		EnrollDeviceRequest{DisplayName: "Teddy Mac"},
	)
	if err == nil || err.Error() != "principal_id is required" {
		t.Fatalf("err = %v", err)
	}
}

func TestNormalizeDeviceEnrollmentRejectsMissingWorkspace(t *testing.T) {
	_, err := normalizeDeviceEnrollment(
		AuthorizationResult{PrincipalID: "user-1"},
		" ",
		EnrollDeviceRequest{DisplayName: "Teddy Mac"},
	)
	if err == nil || err.Error() != "workspace_id is required" {
		t.Fatalf("err = %v", err)
	}
}
