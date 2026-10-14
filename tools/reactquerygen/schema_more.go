package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func additionalPropertiesType(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var allowed bool
	if err := json.Unmarshal(raw, &allowed); err == nil {
		if allowed {
			return "unknown", true
		}
		return "", false
	}
	var nested schema
	if err := json.Unmarshal(raw, &nested); err != nil {
		return "unknown", true
	}
	return schemaType(nested, true), true
}

func stringUnion(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, " | ")
}

func typeDescription(name string, s schema) string {
	if description := strings.TrimSpace(s.Description); description != "" {
		return description
	}
	return name + " 타입입니다."
}
