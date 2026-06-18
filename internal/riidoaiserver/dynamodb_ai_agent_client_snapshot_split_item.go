package riidoaiserver

import (
	"encoding/json"
	"strconv"
	"strings"
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

func dynamoDBAIAgentClientSnapshotFromManifest(item map[string]map[string]string) AIAgentClientSnapshot {
	return AIAgentClientSnapshot{
		SchemaVersion:           dynamoDBStringValue(item, "schema_version"),
		SavedAt:                 parseDynamoDBAIAgentClientSnapshotSavedAt(item),
		WorkspaceID:             dynamoDBStringValue(item, "workspace_id"),
		NextDeviceCredentialSeq: parseDynamoDBAIAgentClientSnapshotInt(item, "next_device_credential_seq"),
		NextDaemonCommand:       parseDynamoDBAIAgentClientSnapshotInt(item, "next_daemon_command"),
	}
}

func dynamoDBAIAgentClientSnapshotCurrentItem(items []map[string]map[string]string) map[string]map[string]string {
	for _, item := range items {
		if dynamoDBStringValue(item, "sk") == dynamoDBAIAgentClientSnapshotSK {
			return item
		}
	}
	return nil
}

func dynamoDBAIAgentClientSnapshotItemsByPart(items []map[string]map[string]string) map[string]map[string]map[string]string {
	out := make(map[string]map[string]map[string]string, len(items))
	for _, item := range items {
		partName := dynamoDBAIAgentClientSnapshotPartName(item)
		if partName != "" {
			out[partName] = item
		}
	}
	return out
}

func dynamoDBAIAgentClientSnapshotPartHashes(items []map[string]map[string]string) map[string]string {
	out := map[string]string{}
	for _, item := range items {
		partName := dynamoDBStringValue(item, "part_name")
		if partName == "" {
			continue
		}
		if hash := dynamoDBStringValue(item, "part_hash"); hash != "" {
			out[partName] = hash
		}
	}
	return out
}

func dynamoDBAIAgentClientSnapshotPartName(item map[string]map[string]string) string {
	if partName := dynamoDBStringValue(item, "part_name"); partName != "" {
		return partName
	}
	sk := dynamoDBStringValue(item, "sk")
	if !strings.HasPrefix(sk, "PART#") {
		return ""
	}
	return strings.TrimPrefix(sk, "PART#")
}

func dynamoDBAIAgentClientSnapshotItemIsLegacy(item map[string]map[string]string) bool {
	return dynamoDBStringValue(item, "snapshot_gzip") != "" || dynamoDBStringValue(item, "snapshot_json") != ""
}

func dynamoDBAIAgentClientSnapshotPartSK(name string) string {
	return "PART#" + name
}

func mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes map[string]string) string {
	raw, err := json.Marshal(hashes)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func parseDynamoDBAIAgentClientSnapshotSavedAt(item map[string]map[string]string) time.Time {
	savedAt, _ := time.Parse(time.RFC3339Nano, dynamoDBStringValue(item, "saved_at"))
	return savedAt
}

func parseDynamoDBAIAgentClientSnapshotInt(item map[string]map[string]string, key string) int {
	value, _ := strconv.Atoi(dynamoDBNumberValue(item, key))
	return value
}
