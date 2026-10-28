package riidoaiserver

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

const (
	conversationChunkPK    = "AI_AGENT_CLIENT#history-chunks"
	conversationChunkBytes = 192 << 10
)

// Parts remain one atomic manifest update. Large immutable payloads are staged
// separately, so neither the item nor transaction size grows with the history.
func conversationPartItems(snapshot AIAgentClientSnapshot, part dynamoDBAIAgentClientSnapshotPart) ([]map[string]map[string]string, error) {
	item := dynamoDBAIAgentClientSnapshotPartItem(snapshot, part)
	items := []map[string]map[string]string{}
	if snapshotHasConversationHistory(snapshot) && (part.name == dynamoDBAIAgentClientSnapshotPartTaskThreadMessages || part.name == dynamoDBAIAgentClientSnapshotPartTaskThreads) && len(part.gzip) > conversationChunkBytes {
		hashes := []string{}
		for offset := 0; offset < len(part.gzip); offset += conversationChunkBytes {
			body := part.gzip[offset:min(offset+conversationChunkBytes, len(part.gzip))]
			hash := sha256Hex([]byte(body))
			hashes = append(hashes, hash)
			items = append(items, map[string]map[string]string{
				"pk": {"S": conversationChunkPK}, "sk": {"S": hash}, "body": {"S": body},
			})
		}
		encoded, err := json.Marshal(hashes)
		if err != nil {
			return nil, err
		}
		delete(item, "part_gzip")
		item["chunk_hashes_json"] = map[string]string{"S": string(encoded)}
	}
	encoded, err := json.Marshal(item)
	if err != nil || len(encoded) > 380<<10 {
		return nil, errors.New("snapshot part exceeds DynamoDB item budget")
	}
	return append(items, item), nil
}

func conversationChunkHashes(item map[string]map[string]string) ([]string, error) {
	value := dynamoDBStringValue(item, "chunk_hashes_json")
	if value == "" {
		return nil, nil
	}
	var hashes []string
	if json.Unmarshal([]byte(value), &hashes) != nil || len(hashes) == 0 {
		return nil, errors.New("invalid conversation chunk manifest")
	}
	for _, hash := range hashes {
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != 32 {
			return nil, errors.New("invalid conversation chunk identity")
		}
	}
	return hashes, nil
}
