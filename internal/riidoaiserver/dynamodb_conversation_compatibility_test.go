package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestConversationManifestAcceptsLegacyAndFencesOldWriter(t *testing.T) {
	legacy := snapshotTestRecord(fixedSnapshotTestNow())
	for _, snapshot := range []AIAgentClientSnapshot{legacy, conversationSnapshotForTest(1)} {
		items, _, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items); err != nil {
			t.Fatalf("current reader cannot reopen: %v", err)
		}
		raw := []byte(dynamoDBStringValue(dynamoDBAIAgentClientSnapshotCurrentItem(items), "part_hashes_json"))
		var oldHashes map[string]string
		oldErr := json.Unmarshal(raw, &oldHashes)
		if snapshotHasConversationHistory(snapshot) != (oldErr != nil) {
			t.Fatalf("legacy writer fence: %v", oldErr)
		}
	}
}
