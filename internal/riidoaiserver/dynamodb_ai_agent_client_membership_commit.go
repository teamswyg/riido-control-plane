package riidoaiserver

import (
	"fmt"
)

func decodeCommittedTaskAgentMemberships(current, item map[string]map[string]string, snapshot *AIAgentClientSnapshot) error {
	return decodeCommittedAIAgentClientPart(dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships, current, item, nil, snapshot)
}

func decodeCommittedAIAgentClientPart(name string, current, item map[string]map[string]string, items []map[string]map[string]string, snapshot *AIAgentClientSnapshot) error {
	var hashes map[string]string
	manifest := dynamoDBStringValue(current, "part_hashes_json")
	if manifest == "" {
		return nil // Legacy CURRENT has no membership commit; reconstruct from threads.
	}
	if err := decodeConversationManifestHashes([]byte(manifest), &hashes); err != nil {
		return fmt.Errorf("decode DynamoDB AI Agent client membership manifest: %w", err)
	}
	expected, advertised := hashes[name]
	if !advertised {
		return nil // Ignore leftovers from failed saves or older-server commits.
	}
	if expected == "" || item == nil || dynamoDBStringValue(item, "part_hash") != expected {
		return fmt.Errorf("decode DynamoDB AI Agent client snapshot part %s: committed hash mismatch or missing part", name)
	}
	raw, err := snapshotPartRaw(item, items)
	if err != nil {
		return fmt.Errorf("decode DynamoDB AI Agent client snapshot part %s gzip: %w", name, err)
	}
	if sha256Hex(raw) != expected {
		return fmt.Errorf("decode DynamoDB AI Agent client snapshot part %s: payload hash mismatch", name)
	}
	return unmarshalDynamoDBAIAgentClientSnapshotPart(name, raw, snapshot)
}
