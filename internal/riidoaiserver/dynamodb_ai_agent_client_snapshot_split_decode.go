package riidoaiserver

import (
	"encoding/json"
	"errors"
	"fmt"
)

func decodeSplitDynamoDBAIAgentClientSnapshot(items []map[string]map[string]string) (AIAgentClientSnapshot, int64, error) {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	if current == nil {
		return AIAgentClientSnapshot{}, 0, errors.New("decode DynamoDB AI Agent client split snapshot: CURRENT item is required")
	}
	if dynamoDBStringValue(current, "storage_version") != dynamoDBAIAgentClientSnapshotSplitStorageVersion {
		return AIAgentClientSnapshot{}, 0, fmt.Errorf("unsupported AI Agent client snapshot storage_version %q", dynamoDBStringValue(current, "storage_version"))
	}
	// A consistent Query is not a multi-item snapshot. Reject mixed parts
	// observed across a concurrent atomic write before publishing any state.
	hashes := map[string]string{}
	if err := json.Unmarshal([]byte(dynamoDBStringValue(current, "part_hashes_json")), &hashes); err != nil {
		return AIAgentClientSnapshot{}, 0, err
	}
	parts := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	for name, hash := range hashes {
		item := parts[name]
		raw, err := gunzipBase64(dynamoDBStringValue(item, "part_gzip"))
		if err != nil || hash != sha256Hex(raw) {
			return AIAgentClientSnapshot{}, 0, errAIAgentSnapshotConflict
		}
	}
	snapshot := dynamoDBAIAgentClientSnapshotFromManifest(current)
	var encodedBytes int64
	byPart := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	for _, name := range dynamoDBAIAgentClientSnapshotRequiredPartNames {
		item := byPart[name]
		if item == nil {
			return AIAgentClientSnapshot{}, 0, fmt.Errorf("decode DynamoDB AI Agent client split snapshot: part %s is required", name)
		}
		encodedBytes += int64(len(dynamoDBStringValue(item, "part_gzip")))
		if err := decodeDynamoDBAIAgentClientSnapshotPart(name, item, &snapshot); err != nil {
			return AIAgentClientSnapshot{}, 0, err
		}
	}
	for _, name := range dynamoDBAIAgentClientSnapshotOptionalPartNames {
		item := byPart[name]
		if item == nil {
			continue
		}
		encodedBytes += int64(len(dynamoDBStringValue(item, "part_gzip")))
		if err := decodeDynamoDBAIAgentClientSnapshotPart(name, item, &snapshot); err != nil {
			return AIAgentClientSnapshot{}, 0, err
		}
	}
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return AIAgentClientSnapshot{}, 0, fmt.Errorf("unsupported AI Agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	return snapshot, encodedBytes, nil
}

func decodeDynamoDBAIAgentClientSnapshotPart(name string, item map[string]map[string]string, snapshot *AIAgentClientSnapshot) error {
	raw, err := gunzipBase64(dynamoDBStringValue(item, "part_gzip"))
	if err != nil {
		return fmt.Errorf("decode DynamoDB AI Agent client snapshot part %s gzip: %w", name, err)
	}
	return unmarshalDynamoDBAIAgentClientSnapshotPart(name, raw, snapshot)
}
