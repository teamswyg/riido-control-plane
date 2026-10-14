package riidoaiserver

import "strings"

func taskMessageContextText(body string, limit int) string {
	trimmed := strings.TrimSpace(body)
	lower := strings.ToLower(trimmed)
	for _, prefix := range []string{"<p", "<div", "<h1", "<h2", "<h3", "<ul", "<ol", "<pre", "<table", "<blockquote"} {
		if strings.HasPrefix(lower, prefix+">") || strings.HasPrefix(lower, prefix+" ") {
			return taskContextText(body, "html", limit)
		}
	}
	return taskContextText(body, "markdown", limit)
}
