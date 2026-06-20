package main

import "strings"

func backtickList(values []string) string {
	if len(values) == 0 {
		return "`-`"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, "`"+value+"`")
	}
	return strings.Join(out, ", ")
}

func sourceLimitationsByID(items []toolLimitation) map[string]toolLimitation {
	out := map[string]toolLimitation{}
	for _, item := range items {
		out[item.ID] = item
	}
	return out
}

func sourceEntriesByID(items []sourceEntry) map[string]sourceEntry {
	out := map[string]sourceEntry{}
	for _, item := range items {
		out[item.NodeID] = item
	}
	return out
}
