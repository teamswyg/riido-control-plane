package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestSplitMembershipRequiresCommittedManifest(t *testing.T) {
	for _, mode := range []string{"unadvertised", "missing", "mismatched", "corrupt-payload", "malformed-manifest", "committed"} {
		t.Run(mode, func(t *testing.T) {
			snapshot := snapshotTestRecord(fixedSnapshotTestNow())
			snapshot.TaskAgentMemberships = map[string]map[string]AIAgentTaskMembership{"task": {"agent-a": {AgentID: "agent-a"}}}
			items, hashes, _, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, nil)
			if err != nil {
				t.Fatal(err)
			}
			current := dynamoDBAIAgentClientSnapshotCurrentItem(items)
			part := dynamoDBAIAgentClientSnapshotItemsByPart(items)[dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships]
			switch mode {
			case "unadvertised":
				delete(hashes, dynamoDBAIAgentClientSnapshotPartTaskAgentMemberships)
				current["part_hashes_json"] = map[string]string{"S": mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes)}
				part["part_gzip"] = map[string]string{"S": "uncommitted invalid bytes"}
			case "missing":
				delete(part, "part_name")
				delete(part, "sk")
			case "mismatched":
				part["part_hash"] = map[string]string{"S": "uncommitted-hash"}
			case "corrupt-payload":
				encoded, err := gzipBase64([]byte(`{}`))
				if err != nil {
					t.Fatal(err)
				}
				part["part_gzip"] = map[string]string{"S": encoded}
			case "malformed-manifest":
				current["part_hashes_json"] = map[string]string{"S": "{"}
			}
			decoded, _, err := decodeSplitDynamoDBAIAgentClientSnapshot(items)
			if mode != "unadvertised" && mode != "committed" {
				if err == nil {
					t.Fatal("accepted uncommitted or corrupt membership")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "unadvertised" && decoded.TaskAgentMemberships != nil {
				t.Fatal("unadvertised stale membership suppressed legacy reconstruction")
			}
			if mode == "committed" {
				got, _ := json.Marshal(decoded.TaskAgentMemberships)
				want, _ := json.Marshal(snapshot.TaskAgentMemberships)
				if string(got) != string(want) {
					t.Fatalf("committed membership changed: %s", got)
				}
			}
		})
	}
}
