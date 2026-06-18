package riidoaiserver

import (
	"strconv"
	"time"
)

func dynamoDBAIAgentClientSnapshotManifestItem(snapshot AIAgentClientSnapshot, hashes map[string]string) map[string]map[string]string {
	return map[string]map[string]string{
		"pk":                         {"S": dynamoDBAIAgentClientSnapshotPK},
		"sk":                         {"S": dynamoDBAIAgentClientSnapshotSK},
		"schema_version":             {"S": AIAgentClientPersistenceSchemaVersion},
		"storage_version":            {"S": dynamoDBAIAgentClientSnapshotSplitStorageVersion},
		"saved_at":                   {"S": snapshot.SavedAt.UTC().Format(time.RFC3339Nano)},
		"workspace_id":               {"S": snapshot.WorkspaceID},
		"next_device_credential_seq": {"N": strconv.Itoa(snapshot.NextDeviceCredentialSeq)},
		"next_daemon_command":        {"N": strconv.Itoa(snapshot.NextDaemonCommand)},
		"part_hashes_json":           {"S": mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes)},
	}
}

func dynamoDBAIAgentClientSnapshotPartItem(snapshot AIAgentClientSnapshot, part dynamoDBAIAgentClientSnapshotPart) map[string]map[string]string {
	return map[string]map[string]string{
		"pk":              {"S": dynamoDBAIAgentClientSnapshotPK},
		"sk":              {"S": dynamoDBAIAgentClientSnapshotPartSK(part.name)},
		"schema_version":  {"S": AIAgentClientPersistenceSchemaVersion},
		"storage_version": {"S": dynamoDBAIAgentClientSnapshotSplitStorageVersion},
		"part_name":       {"S": part.name},
		"part_hash":       {"S": part.hash},
		"part_gzip":       {"S": part.gzip},
		"saved_at":        {"S": snapshot.SavedAt.UTC().Format(time.RFC3339Nano)},
	}
}

func parseDynamoDBAIAgentClientSnapshotSavedAt(item map[string]map[string]string) time.Time {
	savedAt, _ := time.Parse(time.RFC3339Nano, dynamoDBStringValue(item, "saved_at"))
	return savedAt
}

func parseDynamoDBAIAgentClientSnapshotInt(item map[string]map[string]string, key string) int {
	value, _ := strconv.Atoi(dynamoDBNumberValue(item, key))
	return value
}
