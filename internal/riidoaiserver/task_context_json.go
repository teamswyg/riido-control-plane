package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
)

type taskContextJSONNode struct {
	Type    string                     `json:"type"`
	Text    *string                    `json:"text"`
	Attrs   map[string]json.RawMessage `json:"attrs"`
	Content []json.RawMessage          `json:"content"`
	Marks   []taskContextJSONMark      `json:"marks"`
}

type taskContextJSONMark struct {
	Type  string                     `json:"type"`
	Attrs map[string]json.RawMessage `json:"attrs"`
}
type taskContextJSONRenderer struct {
	out      strings.Builder
	nodes    int
	complete bool
}

func readableTaskContextJSON(raw string) (string, bool, error) {
	if len(raw) > 1<<20 {
		return "", false, errors.New("document projection exceeds bound")
	}
	var root map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &root) != nil {
		return "", false, errors.New("invalid document projection")
	}
	if wrapped, ok := root["default"]; ok {
		raw = string(wrapped)
	}
	var node taskContextJSONNode
	if json.Unmarshal([]byte(raw), &node) != nil || node.Type != "doc" {
		return "", false, errors.New("unsupported document projection")
	}
	r := taskContextJSONRenderer{complete: true}
	if err := r.walk([]byte(raw), 0); err != nil {
		return "", false, err
	}
	return strings.TrimSpace(r.out.String()), r.complete, nil
}

func taskContextJSONAttribute(attrs map[string]json.RawMessage, name string) string {
	var value string
	_ = json.Unmarshal(attrs[name], &value)
	return value
}
