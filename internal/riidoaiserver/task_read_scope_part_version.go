package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func readScopePartVersion(kind AssignmentContentKind, id, team, part, source, content string, selector taskReadPartSelector) string {
	encoded, _ := json.Marshal(selector)                                                                                                 //nolint:errchkjson // Fixed scalar selector cannot fail JSON encoding.
	canonical, _ := json.Marshal([]string{"riido-read-part.v1", string(kind), id, team, part, string(encoded), source, "json", content}) //nolint:errchkjson // A string slice cannot fail JSON encoding.
	digest := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(digest[:])
}
