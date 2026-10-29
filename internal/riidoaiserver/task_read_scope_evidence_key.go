package riidoaiserver

import "encoding/json"

func readEvidenceKey(values ...any) string {
	encoded, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func (index *taskReadEvidenceIndex) add(key string, value any) error {
	if _, exists := index.values[key]; key == "" || exists {
		return taskContextReadError{"invalid"}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return taskContextReadError{"invalid"}
	}
	index.values[key] = data
	return nil
}

func readEvidenceBodyKey(kind AssignmentContentKind, id, part, version string, offset any, selector *taskReadPartSelector) (string, error) {
	normalized, err := normalizeReadScopePartSelector(kind, part, selector)
	if err != nil {
		return "", err
	}
	if readScopeBodyPart(part) != "" {
		return readEvidenceKey("body", kind, id, part, version, offset), nil
	}
	return readEvidenceKey("body", kind, id, part, version, offset, normalized), nil
}
