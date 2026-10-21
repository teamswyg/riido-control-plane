package riidoaiserver

import (
	"fmt"
	"strings"
)

func readableTaskContextTiptapJSON(raw []byte) (string, error) {
	document, err := decodeTaskContextTiptapDocument(raw)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	remaining := 20000
	var walk func(taskContextTiptapNode, int) error
	walk = func(node taskContextTiptapNode, depth int) error {
		remaining--
		if depth > 128 || remaining < 0 || node.Type == "" {
			return fmt.Errorf("provider JSON structure exceeds supported bounds")
		}
		switch node.Type {
		case "text":
			out.WriteString(node.Text)
		case "hardBreak":
			out.WriteByte('\n')
		case "image", "imageBlock":
			out.WriteString("[Image: " + node.Attrs.Alt + "]")
			if target := safeTaskContextLink(node.Attrs.Src); target != "" {
				out.WriteString(" (" + target + ")")
			}
		case "customVideo":
			out.WriteString("[Video]")
			if target := safeTaskContextLink(node.Attrs.Src); target != "" {
				out.WriteString(" (" + target + ")")
			}
		case "mention", "riidoMention", "mentionedComponent":
			out.WriteString(taskContextMentionLabel(node))
		case "listItem", "taskItem":
			out.WriteString("\n- ")
		}
		for _, child := range node.Content {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		for _, mark := range node.Marks {
			if mark.Type == "link" {
				if target := safeTaskContextLink(mark.Attrs.Href); target != "" {
					out.WriteString(" (" + target + ")")
				}
			}
		}
		switch node.Type {
		case "paragraph", "heading", "blockquote", "codeBlock", "listItem", "taskItem", "tableRow":
			out.WriteByte('\n')
		case "tableCell", "tableHeader":
			out.WriteString(" | ")
		}
		return nil
	}
	if err := walk(document, 0); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
