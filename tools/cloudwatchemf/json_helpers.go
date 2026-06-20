package main

import "encoding/json"

func keys(raw map[string]json.RawMessage) map[string]bool {
	out := make(map[string]bool, len(raw))
	for key := range raw {
		out[key] = true
	}
	return out
}

func lenRawArray(body json.RawMessage) int {
	var rows []json.RawMessage
	if err := json.Unmarshal(body, &rows); err != nil {
		return 0
	}
	return len(rows)
}

func hasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
