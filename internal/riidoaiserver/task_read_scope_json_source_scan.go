package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

func scanReadScopeJSONSources(data json.RawMessage) (map[readScopeJSONSourceKey][]readScopeJSONSourceSpan, error) {
	invalid := taskContextReadError{"invalid"}
	if !utf8.Valid(data) {
		return nil, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, invalid
	}
	ranges := map[readScopeJSONSourceKey][]readScopeJSONSourceSpan{}
	seen := map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		name, ok := token.(string)
		if err != nil || !ok || seen[name] {
			return nil, invalid
		}
		seen[name] = true
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil {
			return nil, invalid
		}
		kind, array := "", true
		switch name {
		case "links", "github":
			kind = "link"
		case "figma":
			kind, array = "link", false
		case "attachments":
			kind = "attachment"
		case "items":
			kind = "comment"
		}
		if kind != "" && addReadScopeJSONSourceRanges(ranges, raw, int(decoder.InputOffset())-len(raw), kind, array) != nil {
			return nil, invalid
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return nil, invalid
	}
	var trailing json.RawMessage
	if decoder.Decode(&trailing) != io.EOF {
		return nil, invalid
	}
	return ranges, nil
}
