package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDeviceListExplainsInstalledDaemonUpdateRequirement(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	store.devices = []DeviceRecord{{
		DeviceID: "device-a", OwnerPrincipalID: "owner-a", DesktopAppVersion: "0.0.15",
		ConnectedWorkspaceIDs: []string{"workspace-a"},
	}}
	store.daemons = map[string]DeviceDaemonRecord{}
	daemon := DeviceDaemonRecord{
		DeviceID: "device-a", DaemonID: "daemon-a", Profile: "production",
		OwnerPrincipalID: "owner-a", AppVersion: "riido-agentd v0.0.67",
		Availability: DaemonAvailabilityOnline, LastSeenAt: time.Now().UTC(),
	}
	store.putDaemonLocked(daemon)
	store.mu.Unlock()

	response, err := store.ListAIAgentDevices(context.Background(), AuthorizationResult{
		PrincipalID: "workspace-member", WorkspaceID: "workspace-a",
	})
	if err != nil || len(response.Devices) != 1 {
		t.Fatalf("ListAIAgentDevices = %+v, %v", response, err)
	}
	device := response.Devices[0]
	if device.IsOwnedByViewer || device.ClientStatus == nil {
		t.Fatalf("device ownership/status = %+v", device)
	}
	status := device.ClientStatus
	if status.DesktopAppVersion != "0.0.15" || status.DaemonVersion != daemon.AppVersion ||
		status.AgentCapability != DeviceAgentCapabilityUpdateRequired || !status.UpdateRequired || status.AgentSupported {
		t.Fatalf("client status = %+v", status)
	}
}
