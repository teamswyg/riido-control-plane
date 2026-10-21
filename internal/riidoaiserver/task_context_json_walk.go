package riidoaiserver

import (
	"encoding/json"
	"errors"
)

func (r *taskContextJSONRenderer) walk(raw json.RawMessage, depth int) error {
	r.nodes++
	if depth > 64 || r.nodes > 4096 {
		return errors.New("document projection complexity exceeds bound")
	}
	var n taskContextJSONNode
	if json.Unmarshal(raw, &n) != nil || n.Type == "" {
		return errors.New("invalid document node")
	}
	r.prefix(n)
	if n.Type == "text" {
		if n.Text == nil {
			return errors.New("invalid text node")
		}
		r.out.WriteString(*n.Text)
	}
	for _, child := range n.Content {
		if err := r.walk(child, depth+1); err != nil {
			return err
		}
	}
	r.suffix(n)
	return nil
}
