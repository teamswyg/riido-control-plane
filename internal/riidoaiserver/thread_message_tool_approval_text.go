package riidoaiserver

import (
	"strings"

	"golang.org/x/net/html"
)

func toolApprovalMessageText(body string) string {
	ids, err := commentMentionIDs(body)
	if err != nil || len(ids) == 0 {
		return readableTaskContextHTML(body)
	}
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return readableTaskContextHTML(body)
	}
	valid := make(map[string]bool, len(ids))
	for _, id := range ids {
		valid[id] = true
	}
	var stripMentions func(*html.Node)
	stripMentions = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "pre", "code", "blockquote", "template", "script", "style", "noscript":
				return
			}
			if hiddenContextHTML(node) {
				return
			}
		}
		for child := node.FirstChild; child != nil; {
			next := child.NextSibling
			attrs := map[string]string{}
			for _, attr := range child.Attr {
				attrs[attr.Key] = attr.Val
			}
			if child.Type == html.ElementNode && child.Data == "span" &&
				attrs["data-type"] == "agent-mention" && attrs["data-mention-kind"] == "agent" &&
				valid[attrs["data-mention-id"]] {
				node.RemoveChild(child)
			} else {
				stripMentions(child)
			}
			child = next
		}
	}
	stripMentions(root)
	var out strings.Builder
	if err := html.Render(&out, root); err != nil {
		return readableTaskContextHTML(body)
	}
	return readableTaskContextHTML(out.String())
}
