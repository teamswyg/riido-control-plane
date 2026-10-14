package riidoaiserver

import "testing"

func assertSnapshotManifestAndParts(t *testing.T, items map[string]map[string]map[string]string, wantCredentialSeq, wantDaemonCommand string) {
	t.Helper()
	current := items[dynamoDBAIAgentClientSnapshotSK]
	if current == nil {
		t.Fatal("missing CURRENT manifest item")
	}
	assertDynamoDBString(t, current, "pk", dynamoDBAIAgentClientSnapshotPK)
	assertDynamoDBString(t, current, "sk", dynamoDBAIAgentClientSnapshotSK)
	assertDynamoDBString(t, current, "schema_version", AIAgentClientPersistenceSchemaVersion)
	assertDynamoDBString(t, current, "storage_version", dynamoDBAIAgentClientSnapshotSplitStorageVersion)
	assertDynamoDBNumber(t, current, "next_device_credential_seq", wantCredentialSeq)
	assertDynamoDBNumber(t, current, "next_daemon_command", wantDaemonCommand)
	if current["snapshot_gzip"]["S"] != "" || current["snapshot_json"]["S"] != "" {
		t.Fatalf("CURRENT item should be a small manifest, got %+v", current)
	}
	for _, partName := range dynamoDBAIAgentClientSnapshotPartNames {
		assertSnapshotPart(t, items, partName)
	}
}

func assertSnapshotPart(t *testing.T, items map[string]map[string]map[string]string, partName string) {
	t.Helper()
	item := items[dynamoDBAIAgentClientSnapshotPartSK(partName)]
	if item == nil {
		t.Fatalf("missing split part %s", partName)
	}
	if item["part_gzip"]["S"] == "" || item["part_hash"]["S"] == "" {
		t.Fatalf("part %s missing gzip/hash: %+v", partName, item)
	}
}
