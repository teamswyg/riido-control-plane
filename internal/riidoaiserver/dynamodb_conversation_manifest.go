package riidoaiserver

import (
	"encoding/json"
	"errors"
)

const conversationHistoryManifestGuard = "_conversation_history_format"

// The non-string format marker also fences old membership-owned writers, which
// decode CURRENT as map[string]string before saving. They cannot trim history
// and publish a successful replacement of this manifest.
func conversationManifestHashes(snapshot AIAgentClientSnapshot, hashes map[string]string) string {
	if !snapshotHasConversationHistory(snapshot) {
		return mustJSONDynamoDBAIAgentClientSnapshotPartHashes(hashes)
	}
	values := make(map[string]any, len(hashes)+1)
	for key, hash := range hashes {
		values[key] = hash
	}
	values[conversationHistoryManifestGuard] = 1
	body, _ := json.Marshal(values)
	return string(body)
}

func decodeConversationManifestHashes(body []byte, hashes *map[string]string) error {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(body, &values); err != nil {
		return err
	}
	if format, found := values[conversationHistoryManifestGuard]; found {
		if string(format) != "1" {
			return errors.New("unsupported conversation history format")
		}
		delete(values, conversationHistoryManifestGuard)
	}
	*hashes = make(map[string]string, len(values))
	for key, value := range values {
		var hash string
		if err := json.Unmarshal(value, &hash); err != nil {
			return err
		}
		(*hashes)[key] = hash
	}
	return nil
}
