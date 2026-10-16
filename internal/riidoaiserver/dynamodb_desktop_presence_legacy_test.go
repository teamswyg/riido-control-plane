package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDesktopPresenceReadsLegacyInlineAndRejectsMissingCommittedPart(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	snapshot := snapshotTestRecord(fixedSnapshotTestNow())
	snapshot.DeviceConnectionGrants = []DeviceConnectionGrant{{DeviceID: "device", PrincipalID: "owner", WorkspaceID: "workspace", ConnectedAt: snapshot.SavedAt}}
	raw, _ := json.Marshal(snapshot)
	current := map[string]map[string]string{"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": dynamoDBAIAgentClientSnapshotSK}, "snapshot_json": {"S": string(raw)}}
	fixture.mu.Lock()
	fixture.items[membershipDynamoKey(current)] = current
	fixture.mu.Unlock()
	store := newSnapshotDynamoDBStore(t, snapshot.SavedAt, fixture.server.URL, nil)
	defer store.Close()
	grants, err := store.loadDeviceConnectionGrants(context.Background())
	if err != nil || len(grants) != 1 {
		t.Fatalf("legacy grant missing: %v", err)
	}
	seen, err := desktopPresenceFromGrants(grants, "owner")
	if err != nil || !seen.Equal(snapshot.SavedAt) {
		t.Fatalf("legacy timestamp not normalized: %v", err)
	}
	modern := dynamoDBAIAgentClientSnapshotManifestItem(snapshot, map[string]string{dynamoDBAIAgentClientSnapshotPartDeviceConnections: "hash-required"})
	fixture.mu.Lock()
	fixture.items[membershipDynamoKey(current)] = modern
	fixture.mu.Unlock()
	if _, err := store.loadDeviceConnectionGrants(context.Background()); err == nil {
		t.Fatal("missing advertised part silently treated as no connection")
	}
}
