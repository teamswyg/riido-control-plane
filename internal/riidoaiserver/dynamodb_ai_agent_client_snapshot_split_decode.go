package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

func decodeSplitDynamoDBAIAgentClientSnapshot(items []map[string]map[string]string) (AIAgentClientSnapshot, int64, error) {
	current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
	if current == nil {
		return AIAgentClientSnapshot{}, 0, errors.New("decode DynamoDB AI Agent client split snapshot: CURRENT item is required")
	}
	if dynamoDBStringValue(current, "storage_version") != dynamoDBAIAgentClientSnapshotSplitStorageVersion {
		return AIAgentClientSnapshot{}, 0, fmt.Errorf("unsupported AI Agent client snapshot storage_version %q", dynamoDBStringValue(current, "storage_version"))
	}
	snapshot := dynamoDBAIAgentClientSnapshotFromManifest(current)
	var encodedBytes int64
	byPart := dynamoDBAIAgentClientSnapshotItemsByPart(items)
	for _, name := range dynamoDBAIAgentClientSnapshotPartNames {
		item := byPart[name]
		if item == nil {
			return AIAgentClientSnapshot{}, 0, fmt.Errorf("decode DynamoDB AI Agent client split snapshot: part %s is required", name)
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

func decodeLegacyDynamoDBAIAgentClientSnapshot(item map[string]map[string]string) (AIAgentClientSnapshot, int64, error) {
	snapshotReader, snapshotBytes, err := legacyDynamoDBAIAgentClientSnapshotReader(item)
	if err != nil {
		return AIAgentClientSnapshot{}, 0, err
	}
	snapshot, err := decodeAIAgentClientSnapshot(snapshotReader)
	if err != nil {
		return AIAgentClientSnapshot{}, 0, fmt.Errorf("decode DynamoDB AI Agent client snapshot json: %w", err)
	}
	if snapshot.SchemaVersion != AIAgentClientPersistenceSchemaVersion {
		return AIAgentClientSnapshot{}, 0, fmt.Errorf("unsupported AI Agent client snapshot schema_version %q", snapshot.SchemaVersion)
	}
	return snapshot, snapshotBytes, nil
}

func legacyDynamoDBAIAgentClientSnapshotReader(item map[string]map[string]string) (io.Reader, int64, error) {
	if gzipped := dynamoDBStringValue(item, "snapshot_gzip"); gzipped != "" {
		raw, err := gunzipBase64(gzipped)
		if err != nil {
			return nil, 0, fmt.Errorf("decode DynamoDB AI Agent client snapshot gzip: %w", err)
		}
		return bytes.NewReader(raw), int64(len(gzipped)), nil
	}
	rawSnapshot := dynamoDBStringValue(item, "snapshot_json")
	if rawSnapshot == "" {
		return nil, 0, errors.New("decode DynamoDB AI Agent client snapshot response: snapshot_gzip or snapshot_json is required")
	}
	return strings.NewReader(rawSnapshot), int64(len(rawSnapshot)), nil
}

func decodeDynamoDBAIAgentClientSnapshotPart(name string, item map[string]map[string]string, snapshot *AIAgentClientSnapshot) error {
	raw, err := gunzipBase64(dynamoDBStringValue(item, "part_gzip"))
	if err != nil {
		return fmt.Errorf("decode DynamoDB AI Agent client snapshot part %s gzip: %w", name, err)
	}
	return unmarshalDynamoDBAIAgentClientSnapshotPart(name, raw, snapshot)
}

func unmarshalDynamoDBAIAgentClientSnapshotPart(name string, raw []byte, snapshot *AIAgentClientSnapshot) error {
	switch name {
	case dynamoDBAIAgentClientSnapshotPartDevices:
		return json.Unmarshal(raw, &snapshot.Devices)
	case dynamoDBAIAgentClientSnapshotPartDeviceCredentials:
		return json.Unmarshal(raw, &snapshot.DeviceCredentials)
	case dynamoDBAIAgentClientSnapshotPartDaemons:
		return json.Unmarshal(raw, &snapshot.Daemons)
	case dynamoDBAIAgentClientSnapshotPartAgents:
		return json.Unmarshal(raw, &snapshot.Agents)
	case dynamoDBAIAgentClientSnapshotPartFixtures:
		return json.Unmarshal(raw, &snapshot.Fixtures)
	case dynamoDBAIAgentClientSnapshotPartTaskThreads:
		return json.Unmarshal(raw, &snapshot.TaskThreads)
	case dynamoDBAIAgentClientSnapshotPartEvents:
		return json.Unmarshal(raw, &snapshot.Events)
	default:
		return fmt.Errorf("unknown AI Agent client snapshot part %s", name)
	}
}
