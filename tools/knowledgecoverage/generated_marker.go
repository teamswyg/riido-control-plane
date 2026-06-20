package main

import "strings"

func generatedToolFromMarker(text string) string {
	line := generatedMarkerLine(text)
	if line == "" {
		return ""
	}
	for _, prefix := range []string{"go run ./", "./", ""} {
		if tool := generatedToolAfterPrefix(line, prefix); tool != "" {
			return tool
		}
	}
	return ""
}

func generatedMarkerLine(text string) string {
	start := strings.Index(text, generatedMarker)
	if start < 0 {
		return ""
	}
	line := text[start:]
	if end := strings.Index(line, "-->"); end >= 0 {
		line = line[:end]
	}
	return line
}

func generatedToolAfterPrefix(line, prefixBeforeTool string) string {
	idx := strings.Index(line, prefixBeforeTool+"tools/")
	if idx < 0 {
		return ""
	}
	value := line[idx+len(prefixBeforeTool):]
	for i, r := range value {
		if i == 0 {
			continue
		}
		if r == ' ' || r == ';' || r == '.' {
			return strings.TrimPrefix(value[:i], "./")
		}
	}
	return strings.TrimPrefix(value, "./")
}
