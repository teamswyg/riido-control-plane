package riidoaiserver

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Mirrors the Client AgentMessagePreview HTML detection; Markdown is inert.
var commentHTMLStart = regexp.MustCompile(`(?i)^\s*</?(?:p|div|span|br|strong|b|em|i|u|s|ul|ol|li|code|pre|blockquote|a|h[1-6]|table|thead|tbody|tr|th|td|script|style|iframe|img|svg|math|object|embed|form|input|button|video|audio|source)(?:\s[^>]*)?>`)

// Some runtimes concatenate plain progress paragraphs and the final HTML reply.
// Only completed Agent results may use this envelope. Human messages retain
// their original HTML/Markdown boundary.
func completedAnswerMentionIDs(body string) ([]string, error) {
	if commentHTMLStart.MatchString(body) {
		return commentMentionIDs(body)
	}
	start := finalAnswerParagraph.FindStringSubmatchIndex(body)
	if start == nil {
		return nil, nil
	}
	// An earlier markup container or a Markdown code delimiter can quote the
	// entire suffix. Fail closed when this is not a plain progress prefix.
	prefix := body[:start[0]]
	if strings.ContainsAny(inlineProgressCode.ReplaceAllString(prefix, ""), "<`") || strings.Contains(prefix, "~~~") {
		return nil, nil
	}
	answer := body[start[2]:]
	nodes, err := html.ParseFragment(strings.NewReader(answer), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
			continue
		}
		if n.Type != html.ElementNode {
			return nil, nil
		}
	}
	ids, err := commentMentionIDs(answer)
	if len(ids) > 0 && len(body) > 128<<10 {
		return nil, errors.New("oversized Agent mention")
	}
	return ids, err
}

// Single-line inline code is common in progress updates. Never strip a fence
// or a multiline code span that could contain an apparent answer.
var inlineProgressCode = regexp.MustCompile("`[^`\\r\\n]+`")

var finalAnswerParagraph = regexp.MustCompile(`(?i)\r?\n[ \t]*\r?\n(<p\b)`)

// Only HTML responses use the editor's typed mention contract. Quoted examples
// and Markdown code must not turn documentation into an execution request.
func commentMentionIDs(body string) ([]string, error) {
	if !commentHTMLStart.MatchString(body) {
		return nil, nil
	}
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	var ids []string
	seen := map[string]bool{}
	var walk func(*html.Node) error
	walk = func(node *html.Node) error {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "pre", "code", "blockquote", "script", "style", "template", "noscript":
				return nil
			}
			attrs := map[string]string{}
			for _, a := range node.Attr {
				if _, ok := attrs[a.Key]; !ok {
					attrs[a.Key] = a.Val
				}
			}
			if _, hidden := attrs["hidden"]; hidden || attrs["aria-hidden"] == "true" {
				return nil
			}
			if node.Data == "span" && attrs["data-type"] == "agent-mention" && attrs["data-mention-kind"] == "agent" {
				id := attrs["data-mention-id"]
				if strings.TrimSpace(id) == "" || len(id) > 128 || len(body) > 128<<10 {
					return errors.New("invalid or oversized Agent mention")
				}
				if !seen[id] {
					if len(ids) == 8 {
						return errors.New("too many Agent mentions")
					}
					ids = append(ids, id)
					seen[id] = true
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(doc); err != nil {
		return nil, err
	}
	return ids, nil
}

func commentChildBudget(remaining, count, index int) int {
	budget := (remaining - count) / count
	if index < (remaining-count)%count {
		budget++
	}
	return budget
}
