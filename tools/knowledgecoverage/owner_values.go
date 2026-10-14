package main

import "strings"

func ownerObjectValue(object map[string]any, key string) (map[string]any, bool) {
	var current any = object
	for _, part := range strings.Split(key, ".") {
		next, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = next[part]
		if !ok {
			return nil, false
		}
	}
	value, ok := current.(map[string]any)
	return value, ok
}

func ownerString(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return value
}
