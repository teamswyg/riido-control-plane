package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type assignmentContextCache struct {
	enabled bool
	known   string
}

func assignmentContextCacheRequest(headers http.Header, view string) (assignmentContextCache, error) {
	versions := headers.Values("X-Riido-Context-Cache-Version")
	if view != AssignmentContextViewContent || len(versions) != 1 || versions[0] != "1" {
		return assignmentContextCache{}, nil
	}
	known := headers.Values("X-Riido-Context-Known-Digest")
	cache := assignmentContextCache{enabled: true}
	if len(known) == 0 {
		return cache, nil
	}
	if len(known) != 1 || !validAssignmentContextCacheDigest(known[0]) {
		return assignmentContextCache{}, errors.New("invalid content cache digest")
	}
	cache.known = known[0]
	return cache, nil
}

func validAssignmentContextCacheDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, ch := range value[7:] {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}

func applyAssignmentContextCache(cache assignmentContextCache, agentID string, req AgentAssignmentContextRequest, response *AgentAssignmentContextResponse) error {
	if !cache.enabled || response.View != AssignmentContextViewContent || response.Live {
		return nil
	}
	// Request IDs change audit identity, not representation identity. Never
	// place device credentials or lease secrets into the digest representation.
	representation := struct {
		Version      int                            `json:"version"`
		AgentID      string                         `json:"agent_id"`
		AssignmentID string                         `json:"assignment_id"`
		Target       string                         `json:"target"`
		Cursor       string                         `json:"cursor"`
		MaxBytes     int                            `json:"max_bytes"`
		Response     AgentAssignmentContextResponse `json:"response"`
	}{1, agentID, req.AssignmentID, req.Target, req.Cursor, req.MaxBytes, *response}
	representation.Response.ContentDigest = ""
	representation.Response.NotModified = false
	encoded, err := json.Marshal(representation)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(encoded)
	response.ContentDigest = "sha256:" + hex.EncodeToString(digest[:])
	if cache.known == response.ContentDigest {
		response.NotModified = true
		response.Data = nil
	}
	return nil
}
