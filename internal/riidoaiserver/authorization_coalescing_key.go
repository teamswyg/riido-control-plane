package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func authorizationCoalescingKey(bearerToken string, req AuthorizationRequest) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(bearerToken)))
	parts := []string{
		hex.EncodeToString(hash[:]),
		string(req.Resource),
		string(req.Action),
		strings.TrimSpace(req.WorkspaceID),
		strings.TrimSpace(req.AgentID),
		strings.TrimSpace(req.DeviceID),
		strings.TrimSpace(req.TaskID),
	}
	return strings.Join(parts, "\x00")
}
