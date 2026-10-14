package riidoaiserver

import "encoding/json"

func dynamoDBAIAgentClientSnapshotFromManifest(item map[string]map[string]string) AIAgentClientSnapshot {
	return AIAgentClientSnapshot{
		SchemaVersion:           dynamoDBStringValue(item, "schema_version"),
		SavedAt:                 parseDynamoDBAIAgentClientSnapshotSavedAt(item),
		WorkspaceID:             dynamoDBStringValue(item, "workspace_id"),
		NextDeviceCredentialSeq: parseDynamoDBAIAgentClientSnapshotInt(item, "next_device_credential_seq"),
		NextDaemonCommand:       parseDynamoDBAIAgentClientSnapshotInt(item, "next_daemon_command"),
	}
}

func mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes map[string]string) string {
	raw, err := json.Marshal(hashes)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
