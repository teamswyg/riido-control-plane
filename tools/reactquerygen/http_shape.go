package main

import (
	"sort"
	"strings"
)

func requestType(op operation) string {
	if op.RequestBody == nil {
		return ""
	}
	for _, content := range op.RequestBody.Content {
		return schemaType(content.Schema, true)
	}
	return ""
}

func responseType(op operation) string {
	for _, ok := range successfulResponses(op.Responses) {
		for _, content := range ok.Content {
			return schemaType(content.Schema, true)
		}
	}
	return "unknown"
}

func isEventStream(op operation) bool {
	for _, ok := range successfulResponses(op.Responses) {
		for contentType := range ok.Content {
			if strings.EqualFold(contentType, "text/event-stream") {
				return true
			}
		}
	}
	return false
}

func successfulResponses(responses map[string]response) []response {
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		if len(status) == 3 && status[0] == '2' {
			statuses = append(statuses, status)
		}
	}
	sort.Strings(statuses)
	out := make([]response, 0, len(statuses))
	for _, status := range statuses {
		out = append(out, responses[status])
	}
	return out
}
