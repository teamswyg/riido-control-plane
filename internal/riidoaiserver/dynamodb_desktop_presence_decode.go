package riidoaiserver

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

const desktopPresencePartMaxBytes = 16 << 20

func decodeDesktopPresencePart(item map[string]map[string]string) ([]DeviceConnectionGrant, error) {
	name := dynamoDBAIAgentClientSnapshotPartDeviceConnections
	if dynamoDBStringValue(item, "pk") != dynamoDBAIAgentClientSnapshotPK ||
		dynamoDBStringValue(item, "sk") != dynamoDBAIAgentClientSnapshotPartSK(name) ||
		dynamoDBStringValue(item, "part_name") != name ||
		dynamoDBStringValue(item, "schema_version") != AIAgentClientPersistenceSchemaVersion ||
		dynamoDBStringValue(item, "storage_version") != dynamoDBAIAgentClientSnapshotSplitStorageVersion {
		return nil, errors.New("desktop presence part envelope is invalid")
	}
	raw, err := desktopPresenceGunzip(dynamoDBStringValue(item, "part_gzip"))
	if err != nil {
		return nil, err
	}
	if sha256Hex(raw) != dynamoDBStringValue(item, "part_hash") {
		return nil, errAIAgentSnapshotConflict
	}
	var grants []DeviceConnectionGrant
	if err := json.Unmarshal(raw, &grants); err != nil {
		return nil, err
	}
	return grants, nil
}

func desktopPresenceGunzip(encoded string) ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	raw, err := io.ReadAll(io.LimitReader(gz, desktopPresencePartMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > desktopPresencePartMaxBytes {
		return nil, errors.New("desktop presence part exceeds read limit")
	}
	return raw, nil
}
