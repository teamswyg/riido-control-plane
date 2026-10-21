package riidoaiserver

import (
	"strings"
	"testing"
)

func TestTaskContextJSONRenderPreservesStructuredEvidence(t *testing.T) {
	raw := `{"default":{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Acceptance"}]},{"type":"taskList","content":[{"type":"taskItem","attrs":{"checked":true},"content":[{"type":"paragraph","content":[{"type":"text","text":"Fix JWT"}]}]}]},{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Result"}]}]}]}]},{"type":"paragraph","content":[{"type":"text","text":"safe","marks":[{"type":"link","attrs":{"href":"https://example.test/spec"}}]},{"type":"text","text":"hidden-url","marks":[{"type":"link","attrs":{"href":"https://example.test/?token=PRIVATE_CANARY"}}]}]},{"type":"image","attrs":{"alt":"diagram","src":"https://example.test/img.png"}}]}}`
	text, _, err := readableTaskContextJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Acceptance", "Fix JWT", "Result", "https://example.test/spec", "diagram"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(text, "PRIVATE_CANARY") {
		t.Fatal("credential URL retained")
	}
}

func TestTaskContextJSONInvalidIsNeverAnEmptyDocument(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `{"default":null}`, `{"type":"text","text":"bad-root"}`, `{"type":"doc","content":[{"type":"text"}]}`, `{"type":"doc","content":[null]}`, `{"type":"doc"} trailing`} {
		if _, _, err := readableTaskContextJSON(raw); err == nil {
			t.Fatalf("invalid accepted: %s", raw)
		}
	}
	text, complete, err := readableTaskContextJSON(`{"type":"doc","content":[]}`)
	if err != nil || text != "" || !complete {
		t.Fatal("valid empty document rejected")
	}
}

func TestTaskContextJSONUnsupportedAndComplexityRemainExplicit(t *testing.T) {
	text, complete, err := readableTaskContextJSON(`{"type":"doc","content":[{"type":"futureWidget","content":[{"type":"text","text":"retained"}]}]}`)
	if err != nil || complete || !strings.Contains(text, "retained") || !strings.Contains(text, "Unsupported") {
		t.Fatal("unknown node silently lost")
	}
	raw := `{"type":"paragraph","content":[{"type":"text","text":"nested"}]}`
	for range 65 {
		raw = `{"type":"paragraph","content":[` + raw + `]}`
	}
	if _, _, err := readableTaskContextJSON(`{"type":"doc","content":[` + raw + `]}`); err == nil {
		t.Fatal("deep document accepted")
	}
	if _, _, err := readableTaskContextJSON(strings.Repeat(" ", 1<<20) + `{}`); err == nil {
		t.Fatal("oversized document accepted")
	}
}
