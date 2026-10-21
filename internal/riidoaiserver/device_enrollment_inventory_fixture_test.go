package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func openEnrollmentInventoryStore(t *testing.T, endpoint string) (*PersistentAIAgentClientStore, *DynamoDBAIAgentClientSnapshot) {
	t.Helper()
	native := newSnapshotDynamoDBStore(t, time.Now(), endpoint, nil)
	t.Cleanup(func() { _ = native.Close() })
	store, err := OpenPersistentAIAgentClientStore(context.Background(), NewDevelopmentAIAgentClientStore(), native)
	if err != nil {
		t.Fatal(err)
	}
	return store, native
}

func assertEnrollmentHeartbeatInventory(t *testing.T, native *DynamoDBAIAgentClientSnapshot, deviceID string, heartbeat time.Time) {
	t.Helper()
	current, found, err := native.LoadAIAgentClientSnapshot(context.Background())
	if err != nil || !found {
		t.Fatal("final snapshot missing")
	}
	for _, device := range current.Devices {
		if device.DeviceID == deviceID {
			if !device.DaemonLastSeenAt.Equal(heartbeat) || len(device.Runtimes) != 1 || device.Runtimes[0].RuntimeID != "fresh-runtime-from-other-replica" {
				t.Fatal("enrollment or following ordinary save erased committed heartbeat inventory")
			}
			return
		}
	}
	t.Fatal("enrolled device missing")
}
