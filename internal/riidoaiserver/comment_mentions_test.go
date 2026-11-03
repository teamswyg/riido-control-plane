package riidoaiserver

import (
	"reflect"
	"strings"
	"testing"
)

func mentionHTML(id string) string {
	return `<span data-type="agent-mention" data-mention-kind="agent" data-mention-id="` + id + `">Agent</span>`
}

func TestCommentMentionsOnlyVisibleTypedHTML(t *testing.T) {
	for _, tc := range []struct {
		body string
		want []string
	}{
		{`<p>` + mentionHTML("a") + mentionHTML("a") + mentionHTML("b") + `</p>`, []string{"a", "b"}},
		{`@a <span data-type="mention" data-mention-kind="member" data-mention-id="b">B</span>`, nil},
		{"```html\n" + mentionHTML("a") + "\n```", nil},
		{`<p>example</p><pre>` + mentionHTML("a") + `</pre><blockquote>` + mentionHTML("b") + `</blockquote>`, nil},
		{`<p hidden>` + mentionHTML("a") + `</p><script>` + mentionHTML("b") + `</script>`, nil},
		{`<p>&lt;span data-type="agent-mention" data-mention-kind="agent" data-mention-id="a"&gt;</p>`, nil},
		{`<!-- example -->` + mentionHTML("a"), nil},
		{`<unknown>` + mentionHTML("a") + `</unknown>`, nil},
		{mentionHTML("a&amp;b"), []string{"a&b"}},
	} {
		got, err := commentMentionIDs(tc.body)
		if err != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: %v %v", tc.body, got, err)
		}
	}
	for _, body := range []string{mentionHTML(""), mentionHTML(strings.Repeat("x", 129)), mentionHTML("a") + strings.Repeat("x", 128<<10)} {
		if _, err := commentMentionIDs(body); err == nil {
			t.Fatal("unbounded mention accepted")
		}
	}
}

func TestCompletedAnswerMentionBoundary(t *testing.T) {
	final := "<p>Tests passed.</p>\n<ul><li>Ready</li></ul>\n<p>" + mentionHTML("reviewer") + " Please review.</p>"
	for _, tc := range []struct {
		name, body string
		want       []string
	}{
		{"single newline before final", "Checking.\n" + final, []string{"reviewer"}},
		{"plain progress", "Checking.\n\nVerified `index.html`.\r\n\r\n" + final, []string{"reviewer"}},
		{"fenced example", "Example:\n```html\n\n" + final + "\n```", nil},
		{"unclosed fence", "Example:\n~~~html\n\n" + final, nil},
		{"quoted container", "Example:\n<blockquote>\n\n" + final + "</blockquote>", nil},
		{"indented code", "Example:\n\n    " + final, nil},
		{"more progress", "Checking.\n\n" + final + "\n\nStill working.", nil},
		{"hidden example", "Checking.\n\n<p><code>" + mentionHTML("example") + "</code></p>", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := completedAnswerMentionIDs(tc.body)
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, %v; want %v", got, err, tc.want)
			}
		})
	}
	if _, err := completedAnswerMentionIDs(strings.Repeat("x", 128<<10) + "\n\n" + final); err == nil {
		t.Fatal("oversized envelope accepted")
	}
}

func TestCommentInvocationBudgetBoundsWholeTree(t *testing.T) {
	var visit func(int, int) int
	visit = func(budget, depth int) int {
		if budget == 0 || depth == 3 {
			return 0
		}
		for n := 1; n <= min(8, budget); n++ {
			total := n
			for i := 0; i < n; i++ {
				total += visit(commentChildBudget(budget, n, i), depth+1)
			}
			if total > budget {
				t.Fatalf("budget=%d fanout=%d total=%d", budget, n, total)
			}
		}
		return budget
	}
	visit(8, 0)
}
