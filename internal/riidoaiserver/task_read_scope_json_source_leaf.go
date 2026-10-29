package riidoaiserver

import (
	"bytes"
	"encoding/json"
)

func addReadScopeJSONSourceRanges(ranges map[readScopeJSONSourceKey][]readScopeJSONSourceSpan, raw json.RawMessage, start int, kind string, array bool) error {
	if bytes.Equal(raw, []byte("null")) {
		return nil
	}
	if !array {
		return addReadScopeJSONSourceLeaf(ranges, raw, start, kind)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('[') {
		return taskContextReadError{"invalid"}
	}
	for decoder.More() {
		var leaf json.RawMessage
		if decoder.Decode(&leaf) != nil {
			return taskContextReadError{"invalid"}
		}
		if err := addReadScopeJSONSourceLeaf(ranges, leaf, start+int(decoder.InputOffset())-len(leaf), kind); err != nil {
			return err
		}
	}
	return nil
}

func addReadScopeJSONSourceLeaf(ranges map[readScopeJSONSourceKey][]readScopeJSONSourceSpan, raw json.RawMessage, start int, kind string) error {
	var leaf struct {
		ID   string  `json:"id"`
		Body *string `json:"body"`
	}
	if json.Unmarshal(raw, &leaf) != nil || leaf.ID == "" || kind == "comment" && leaf.Body == nil {
		return taskContextReadError{"invalid"}
	}
	canonical, err := json.Marshal(raw)
	if err != nil {
		return taskContextReadError{"invalid"}
	}
	version := assignmentContextVersion(string(canonical))
	if kind == "comment" {
		version = assignmentContextVersion(*leaf.Body)
	}
	key := readScopeJSONSourceKey{kind, leaf.ID, version}
	ranges[key] = append(ranges[key], readScopeJSONSourceSpan{start, start + len(raw)})
	return nil
}
