package main

import (
	"sort"
	"strings"
)

func chainsByID(chains []chain) map[string]chain {
	out := map[string]chain{}
	for _, item := range chains {
		out[item.ID] = item
	}
	return out
}

func chainSignature(item chain) string {
	parts := []string{item.ID, item.Observation, item.Hypothesis, item.Decision, item.NextLoop}
	parts = append(parts, prefixedValues("claim", item.Claims)...)
	parts = append(parts, prefixedRefs("change", item.Changes)...)
	parts = append(parts, prefixedRefs("verifier", item.Verifiers)...)
	parts = append(parts, prefixedRefs("evidence", item.Evidence)...)
	return strings.Join(parts, "\x00")
}

func prefixedRefs(prefix string, refs []ref) []string {
	values := []string{}
	for _, item := range refs {
		values = append(values, prefix+":"+item.Kind+":"+item.Path)
	}
	sort.Strings(values)
	return values
}

func prefixedValues(prefix string, values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	for i, value := range out {
		out[i] = prefix + ":" + value
	}
	return out
}
