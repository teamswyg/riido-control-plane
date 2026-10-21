package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
)

func decodeTaskContextTiptapDocument(raw json.RawMessage) (taskContextTiptapNode, error) {
	raw = bytes.TrimSpace(raw)
	// PostgreSQL stores the transformer's fragment map as a JSON-encoded string.
	if len(raw) != 0 && raw[0] == '"' {
		var serialized string
		if err := json.Unmarshal(raw, &serialized); err != nil {
			return taskContextTiptapNode{}, err
		}
		raw = json.RawMessage(serialized)
	}
	var fragments struct {
		Default json.RawMessage `json:"default"`
	}
	if err := json.Unmarshal(raw, &fragments); err != nil {
		return taskContextTiptapNode{}, err
	}
	if len(fragments.Default) != 0 {
		raw = fragments.Default
	}
	var document taskContextTiptapNode
	if err := json.Unmarshal(raw, &document); err != nil {
		return taskContextTiptapNode{}, err
	}
	if document.Type != "doc" {
		return taskContextTiptapNode{}, errors.New("provider JSON root must be doc or a default document fragment")
	}
	return document, nil
}
