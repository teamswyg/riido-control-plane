package riidoaiserver

import (
	"strconv"
	"strings"
)

func (r *taskContextJSONRenderer) prefix(n taskContextJSONNode) {
	switch n.Type {
	case "doc", "text":
	case "paragraph", "blockquote", "codeBlock", "tableRow", "bulletList", "orderedList", "taskList", "details", "detailsSummary", "detailsContent":
		r.out.WriteString("\n")
	case "hardBreak":
		r.out.WriteString("\n")
	case "listItem":
		r.out.WriteString("\n- ")
	case "taskItem":
		r.out.WriteString("\n- [")
		if string(n.Attrs["checked"]) == "true" {
			r.out.WriteString("x")
		} else {
			r.out.WriteString(" ")
		}
		r.out.WriteString("] ")
	case "heading":
		level, _ := strconv.Atoi(string(n.Attrs["level"]))
		if level < 1 || level > 6 {
			level = 1
		}
		r.out.WriteString("\n" + strings.Repeat("#", level) + " ")
	case "table", "tableHeader", "tableCell":
	case "horizontalRule":
		r.out.WriteString("\n---\n")
	case "image", "imageBlock", "video", "videoBlock", "customVideo":
		r.out.WriteString("[Media: " + taskContextJSONAttribute(n.Attrs, "alt") + "]")
		if u := safeTaskContextLink(taskContextJSONAttribute(n.Attrs, "src")); u != "" {
			r.out.WriteString(" (" + u + ")")
		}
	case "riidoMention", "mentionedComponent":
		r.out.WriteString(taskContextRiidoMentionLabel(n))
	case "mention":
		r.out.WriteString("[" + taskContextStandardMentionLabel(n) + "]")
	default:
		r.complete = false
		r.out.WriteString("[Unsupported document element; nested text follows.]")
	}
}

func (r *taskContextJSONRenderer) suffix(n taskContextJSONNode) {
	if n.Type == "text" {
		for _, mark := range n.Marks {
			if mark.Type == "link" {
				if u := safeTaskContextLink(taskContextJSONAttribute(mark.Attrs, "href")); u != "" {
					r.out.WriteString(" (" + u + ")")
				}
			}
		}
	}
	switch n.Type {
	case "tableCell", "tableHeader":
		r.out.WriteString(" | ")
	case "paragraph", "blockquote", "codeBlock", "tableRow", "listItem", "taskItem", "heading", "detailsSummary", "detailsContent":
		r.out.WriteString("\n")
	}
}
