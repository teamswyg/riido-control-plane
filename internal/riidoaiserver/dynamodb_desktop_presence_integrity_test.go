package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"testing"
)

func desktopPresenceTestPart(t *testing.T) map[string]map[string]string {
	t.Helper()
	snapshot := snapshotTestRecord(fixedSnapshotTestNow())
	snapshot.DeviceConnectionGrants = []DeviceConnectionGrant{{DeviceID: "device", PrincipalID: "owner", WorkspaceID: "workspace", LastSeenAt: snapshot.SavedAt}}
	items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	return dynamoDBAIAgentClientSnapshotItemsByPart(items)[dynamoDBAIAgentClientSnapshotPartDeviceConnections]
}

func TestDesktopPresenceRejectsInvalidPartEnvelopeAndPayload(t *testing.T) {
	for _, key := range []string{"pk", "sk", "schema_version", "storage_version", "part_name", "part_hash", "part_gzip"} {
		t.Run(key, func(t *testing.T) {
			part := desktopPresenceTestPart(t)
			part[key] = map[string]string{"S": "private-invalid-value"}
			if _, err := decodeDesktopPresencePart(part); err == nil {
				t.Fatal("invalid part accepted")
			}
		})
	}
	part := desktopPresenceTestPart(t)
	raw := []byte(`{"not":"grant-array"}`)
	part["part_gzip"]["S"], _ = gzipBase64(raw)
	part["part_hash"]["S"] = sha256Hex(raw)
	if _, err := decodeDesktopPresencePart(part); err == nil {
		t.Fatal("non-array payload accepted")
	}
	if _, err := desktopPresenceGunzip(mustPresenceGzip(t, bytes.Repeat([]byte(" "), desktopPresencePartMaxBytes+1))); err == nil {
		t.Fatal("oversized decoded part accepted")
	}
}

func mustPresenceGzip(t *testing.T, raw []byte) string {
	t.Helper()
	value, err := gzipBase64(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestDesktopPresencePartRoundTripDoesNotMutateSnapshotHashes(t *testing.T) {
	part := desktopPresenceTestPart(t)
	before, _ := json.Marshal(part)
	grants, err := decodeDesktopPresencePart(part)
	after, _ := json.Marshal(part)
	if err != nil || len(grants) != 1 || !bytes.Equal(before, after) {
		t.Fatalf("part read changed its input or failed: %v", err)
	}
}
