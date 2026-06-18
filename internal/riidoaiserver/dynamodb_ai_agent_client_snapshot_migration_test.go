package riidoaiserver

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDynamoDBAIAgentClientSnapshotMigratesLegacySnapshotToSplitParts(t *testing.T) {
	fixedNow := fixedSnapshotTestNow()
	legacy := snapshotTestRecord(fixedNow)
	legacy.NextDeviceCredentialSeq = 7
	legacy.NextDaemonCommand = 9
	legacyJSON, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy snapshot: %v", err)
	}
	items := map[string]map[string]map[string]string{
		dynamoDBAIAgentClientSnapshotSK: {
			"pk":             {"S": dynamoDBAIAgentClientSnapshotPK},
			"sk":             {"S": dynamoDBAIAgentClientSnapshotSK},
			"schema_version": {"S": AIAgentClientPersistenceSchemaVersion},
			"snapshot_json":  {"S": string(legacyJSON)},
		},
	}
	fixture := newSnapshotDynamoDBFixture(t, fixedNow, items, nil)
	defer fixture.close()

	loaded, ok, err := fixture.store.LoadAIAgentClientSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadAIAgentClientSnapshot: %v", err)
	}
	if !ok || loaded.NextDeviceCredentialSeq != 7 || loaded.NextDaemonCommand != 9 {
		t.Fatalf("loaded legacy snapshot ok=%v snapshot=%+v", ok, loaded)
	}
	assertSnapshotManifestAndParts(t, items, "7", "9")
}
