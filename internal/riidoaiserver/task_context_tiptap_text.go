package riidoaiserver

import (
	"bytes"
	"encoding/json"
)

func readableTaskContextTiptapJSON(raw []byte) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) != 0 && raw[0] == '"' {
		var serialized string
		if err := json.Unmarshal(raw, &serialized); err != nil {
			return "", err
		}
		raw = []byte(serialized)
	}
	text, _, err := readableTaskContextJSON(string(raw))
	return text, err
}
