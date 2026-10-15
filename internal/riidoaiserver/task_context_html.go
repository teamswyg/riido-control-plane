package riidoaiserver

import (
	"strings"

	"golang.org/x/net/html"
)

func readableTaskContextHTML(content string) string {
	root, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "[Document could not be rendered.]"
	}
	var out strings.Builder
	var walk func(*html.Node, int)
	walk = func(n *html.Node, depth int) {
		if depth > 64 {
			out.WriteString("\n[Deeply nested document content omitted.]\n")
			return
		}
		if n.Type == html.ElementNode && hiddenContextHTML(n) {
			return
		}
		if n.Type == html.TextNode {
			out.WriteString(n.Data)
			return
		}
		writeContextHTMLPrefix(&out, n)
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, depth+1)
		}
		writeContextHTMLSuffix(&out, n)
	}
	walk(root, 0)
	return strings.TrimSpace(out.String())
}

func hiddenContextHTML(n *html.Node) bool {
	switch n.Data {
	case "script", "style", "head", "iframe", "object", "noscript", "template":
		return true
	}
	for _, a := range n.Attr {
		value := strings.ToLower(strings.ReplaceAll(a.Val, " ", ""))
		if a.Key == "hidden" || (a.Key == "aria-hidden" && value == "true") ||
			(a.Key == "style" && (strings.Contains(value, "display:none") || strings.Contains(value, "visibility:hidden"))) {
			return true
		}
	}
	return false
}
