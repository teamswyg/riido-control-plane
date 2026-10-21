package riidoaiserver

import (
	"strings"
	"testing"
)

func TestTaskContextProjectionDoesNotResurrectEmptyDocument(t *testing.T) {
	empty := ""
	jsonEmpty := `{"default":{"type":"doc","content":[]}}`
	for _, projection := range []privateDocumentProjections{{HTML: &empty}, {JSON: &jsonEmpty, HTML: taskContextTestString("<p>OLD</p>")}} {
		got := selectTaskContextDocument(AIAgentTaskContextDocument{Content: "<p>OLD</p>", ContentFormat: "html"}, projection.document())
		if got.Status != "empty" || got.Content != "" || got.AlternateContent != "" {
			t.Fatal("empty projection restored old content")
		}
		var out strings.Builder
		writeTaskContextDocumentBody(&out, got)
		if !strings.Contains(out.String(), "Document is empty") || strings.Contains(out.String(), "OLD") {
			t.Fatal("empty document misreported")
		}
	}
}

func TestTaskContextProjectionJSONOnlyAndConflictsRetainEvidence(t *testing.T) {
	raw := `{"default":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"JSON_ONLY_BODY"}]}]}}`
	got := selectTaskContextDocument(AIAgentTaskContextDocument{Content: "<p>SNAPSHOT</p>", ContentFormat: "html"}, privateDocumentProjections{JSON: &raw}.document())
	if got.Content != "JSON_ONLY_BODY" || got.Source != "provider_json_projection" || !strings.Contains(got.AlternateContent, "SNAPSHOT") {
		t.Fatal("JSON/snapshot evidence lost")
	}
	got = selectTaskContextDocument(AIAgentTaskContextDocument{Content: "NEW_SNAPSHOT"}, privateDocumentProjections{HTML: taskContextTestString("OLD_HTML")}.document())
	if got.Content != "NEW_SNAPSHOT" || got.AlternateContent != "OLD_HTML" {
		t.Fatal("unversioned HTML overwrote component")
	}
	var out strings.Builder
	writeTaskContextDocumentSource(&out, got)
	writeTaskContextDocumentBody(&out, got)
	if !strings.Contains(out.String(), "chronology is unknown") || !strings.Contains(out.String(), "unconfirmed") {
		t.Fatal("freshness inferred")
	}
}

func TestTaskContextProjectionAlternatesShareCurrentAndParentBudgets(t *testing.T) {
	doc := AIAgentTaskContextDocument{Content: strings.Repeat("A", 20000), AlternateContent: strings.Repeat("B", 20000)}
	for _, limit := range []int{taskDocumentContextLimit, parentDocumentContextLimit} {
		got := normalizeTaskContextAlternate(doc, limit)
		if len([]rune(got.Content))+len([]rune(got.AlternateContent)) > limit {
			t.Fatal("alternate exceeded total budget")
		}
	}
}

func taskContextTestString(s string) *string { return &s }
