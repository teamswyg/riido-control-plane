package riidoaiserver

import (
	"encoding/json"
	"strings"
)

type progressMessagePayload struct {
	Code    int            `json:"code"`
	Key     string         `json:"key,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
	Message string         `json:"message,omitempty"`
}

func parseProgressMessagePayload(message string) (progressMessagePayload, bool) {
	message = strings.TrimSpace(message)
	if !strings.HasPrefix(message, "{") {
		return progressMessagePayload{}, false
	}
	var payload progressMessagePayload
	if err := json.Unmarshal([]byte(message), &payload); err != nil || payload.Code <= 0 {
		return progressMessagePayload{}, false
	}
	payload.Key = strings.TrimSpace(payload.Key)
	return payload, true
}
