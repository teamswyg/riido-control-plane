package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func taskThreadAgentSnapshotID(snapshot *AIAgentTaskThreadAgentSnapshot) string {
	if snapshot == nil {
		return ""
	}
	body, err := json.Marshal(snapshot)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(body)
	return "agt_snap_" + hex.EncodeToString(sum[:8])
}
