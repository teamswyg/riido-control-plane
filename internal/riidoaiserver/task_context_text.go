package riidoaiserver

import "strings"

const (
	taskDocumentContextLimit   = 12000
	parentDocumentContextLimit = 2500
	conversationContextLimit   = 6000
	contextTruncationMarker    = "\n[Context truncated: beginning and end retained; omitted content was not read.]\n"
)

func taskContextText(content, format string, limit int) string {
	if strings.EqualFold(strings.TrimSpace(format), "html") {
		content = readableTaskContextHTML(content)
	}
	return boundedTaskContextText(content, limit)
}

func boundedTaskContextText(content string, limit int) string {
	content = strings.TrimSpace(strings.ToValidUTF8(content, "�"))
	runes := []rune(content)
	if len(runes) <= limit {
		return content
	}
	marker := []rune(contextTruncationMarker)
	if limit <= len(marker) {
		return string(runes[:max(limit, 0)])
	}
	remaining := limit - len(marker)
	head := remaining * 2 / 3
	return string(runes[:head]) + contextTruncationMarker + string(runes[len(runes)-(remaining-head):])
}
