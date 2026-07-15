package riidoaiserver

import (
	"testing"
	"time"
)

func TestSplitSnapshotRoundTripsDeviceConnectionGrants(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	snapshot := snapshotTestRecord(now)
	snapshot.DeviceConnectionGrants = []DeviceConnectionGrant{{
		DeviceID:    "device-shared",
		PrincipalID: "account-b",
		WorkspaceID: "workspace-b",
		ConnectedAt: now,
		LastSeenAt:  now,
	}}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatalf("encode split snapshot: %v", err)
	}
	decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
	if err != nil {
		t.Fatalf("decode split snapshot: %v", err)
	}
	if len(decoded.DeviceConnectionGrants) != 1 || decoded.DeviceConnectionGrants[0].PrincipalID != "account-b" {
		t.Fatalf("decoded device connection grants = %+v", decoded.DeviceConnectionGrants)
	}
}
