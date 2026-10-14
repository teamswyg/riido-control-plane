package riidoaiserver

import "fmt"

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
