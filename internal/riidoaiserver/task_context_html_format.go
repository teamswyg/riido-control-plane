package riidoaiserver

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func writeContextHTMLPrefix(out *strings.Builder, n *html.Node) {
	if n.Type != html.ElementNode {
		return
	}
	switch n.Data {
	case "p", "div", "blockquote", "pre", "tr", "ul", "ol":
		out.WriteString("\n")
	case "br":
		out.WriteString("\n")
	case "li":
		out.WriteString("\n- ")
	case "h1", "h2", "h3", "h4", "h5", "h6":
		out.WriteString("\n" + strings.Repeat("#", int(n.Data[1]-'0')) + " ")
	case "img":
		out.WriteString("[Image: " + contextHTMLAttribute(n, "alt") + "]")
	}
}

func writeContextHTMLSuffix(out *strings.Builder, n *html.Node) {
	if n.Type != html.ElementNode {
		return
	}
	switch n.Data {
	case "a", "img":
		key := "href"
		if n.Data == "img" {
			key = "src"
		}
		if target := safeTaskContextLink(contextHTMLAttribute(n, key)); target != "" {
			out.WriteString(" (" + target + ")")
		}
	case "td", "th":
		out.WriteString(" | ")
	case "p", "div", "li", "blockquote", "pre", "tr", "h1", "h2", "h3", "h4", "h5", "h6":
		out.WriteString("\n")
	}
}

func contextHTMLAttribute(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func safeTaskContextLink(raw string) string {
	if len(raw) > 2048 {
		return ""
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return ""
	}
	for key := range u.Query() {
		key = strings.ToLower(key)
		if strings.Contains(key, "token") || strings.Contains(key, "secret") || strings.Contains(key, "signature") || strings.Contains(key, "credential") || key == "key" || key == "api_key" {
			return ""
		}
	}
	return u.String()
}
