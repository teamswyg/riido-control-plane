package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestDeviceListMarksSupportedOwnedDaemonReady(t *testing.T) {
	store := NewDevelopmentAIAgentClientStore()
	store.mu.Lock()
	store.devices = []DeviceRecord{{DeviceID: "device-a", OwnerPrincipalID: "owner-a"}}
	store.daemons = map[string]DeviceDaemonRecord{}
	store.putDaemonLocked(DeviceDaemonRecord{
		DeviceID: "device-a", DaemonID: "daemon-a", OwnerPrincipalID: "owner-a",
		AppVersion: "riido-agentd v0.0.68", Availability: DaemonAvailabilityOnline,
		LastSeenAt: time.Now().UTC(),
	})
	store.mu.Unlock()
	response, err := store.ListAIAgentDevices(context.Background(), AuthorizationResult{PrincipalID: "owner-a"})
	if err != nil || len(response.Devices) != 1 {
		t.Fatalf("ListAIAgentDevices = %+v, %v", response, err)
	}
	device := response.Devices[0]
	if !device.IsOwnedByViewer || device.ClientStatus.AgentCapability != DeviceAgentCapabilityReady ||
		!device.ClientStatus.AgentSupported || device.ClientStatus.UpdateRequired {
		t.Fatalf("ready device = %+v", device)
	}
}
