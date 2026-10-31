package authpep

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
)

func rejectLegacyCLIToken(raw string) bool {
	segments := strings.SplitN(strings.TrimSpace(raw), ".", 4)
	if len(segments) != 3 {
		return false
	}
	// JWT-shaped legacy credentials share the bounded Auth token envelope.
	// Never turn an over-limit classification into fallback authentication.
	if len(raw) > maxJWTAccessTokenBytes {
		return true
	}
	payload, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return false
	}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return false
		}
		// Match Go's claim decoder, including escaped/case-variant keys. Presence
		// of a grant marker owns the boundary even if its value is null or invalid.
		if strings.EqualFold(key, "cliGrantId") {
			return true
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return false
		}
		if strings.EqualFold(key, "clientId") {
			var client string
			if json.Unmarshal(value, &client) == nil && client == "riido-cli" {
				return true
			}
		}
	}
	return false
}
