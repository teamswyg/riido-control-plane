package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func resourceBodyVersion(resource AssignmentScopedResource, part, source, format, content string) string {
	if resource.WorkspaceID == "" || resource.Kind == AssignmentContentMeeting {
		return readScopeBodyVersion(resource.Kind, resource.ID, part, source, format, content)
	}
	return nativeReadScopeDigest([]string{"riido-read-body.v2", resource.WorkspaceID, string(resource.Kind), resource.ID, resource.TeamID, part, source, format, content})
}

func resourcePartVersion(resource AssignmentScopedResource, part, source, content string, selector taskReadPartSelector) string {
	if resource.WorkspaceID == "" || resource.Kind == AssignmentContentMeeting {
		return readScopePartVersion(resource.Kind, resource.ID, resource.TeamID, part, source, content, selector)
	}
	encoded, _ := json.Marshal(selector) //nolint:errchkjson // Fixed scalar selector cannot fail JSON encoding.
	return nativeReadScopeDigest([]string{"riido-read-part.v2", resource.WorkspaceID, string(resource.Kind), resource.ID, resource.TeamID, part, string(encoded), source, "json", content})
}

func nativeReadScopeDigest(identity []string) string {
	canonical, _ := json.Marshal(identity) //nolint:errchkjson // A string slice cannot fail JSON encoding.
	digest := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(digest[:])
}
