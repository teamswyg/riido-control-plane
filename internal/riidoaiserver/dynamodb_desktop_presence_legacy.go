package riidoaiserver

import (
	"context"
	"errors"
)

// Only absent parts need a CURRENT read. Inline snapshots retain their legacy
// connection facts; a split manifest advertising an absent part is an error.
func (s *DynamoDBAIAgentClientSnapshot) readLegacyDesktopPresence(ctx context.Context, credentials AWSCredentials) ([]DeviceConnectionGrant, error) {
	current, err := s.readDesktopPresenceItem(ctx, dynamoDBAIAgentClientSnapshotSK, credentials)
	if err != nil || len(current) == 0 {
		return nil, err
	}
	if dynamoDBStringValue(current, "pk") != dynamoDBAIAgentClientSnapshotPK || dynamoDBStringValue(current, "sk") != dynamoDBAIAgentClientSnapshotSK {
		return nil, errAIAgentSnapshotConflict
	}
	storage := dynamoDBStringValue(current, "storage_version")
	if storage == "" {
		snapshot, err := decodeDesktopPresenceInline(current)
		if err != nil {
			return nil, err
		}
		// Restore also seeds pre-grant snapshots from their original credential.
		restored := NewDevelopmentAIAgentClientStore()
		if err := restored.restoreSnapshot(snapshot); err != nil {
			return nil, err
		}
		return copyDeviceConnectionGrants(restored.deviceConnectionGrants), nil
	}
	if storage != dynamoDBAIAgentClientSnapshotSplitStorageVersion || dynamoDBStringValue(current, "schema_version") != AIAgentClientPersistenceSchemaVersion {
		return nil, errors.New("desktop presence manifest is invalid")
	}
	var hashes map[string]string
	if err := decodeConversationManifestHashes([]byte(dynamoDBStringValue(current, "part_hashes_json")), &hashes); err != nil {
		return nil, err
	}
	if _, present := hashes[dynamoDBAIAgentClientSnapshotPartDeviceConnections]; present {
		return nil, errAIAgentSnapshotConflict
	}
	return nil, nil
}
