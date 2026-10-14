package main

import (
	"strings"
	"testing"
)

func verifyMirroredFigmaAPIGeneratedAnnotationContentPolicy(t *testing.T, policy figmaSourceAPIGeneratedAnnotationContentRule, docText string) {
	t.Helper()
	if policy.CategoryID != "700:0" || policy.CategoryLabel != "API Generated" {
		t.Fatalf("mirrored API Generated annotation content category drifted: %+v", policy)
	}
	if len(policy.LabelFormat) != 3 {
		t.Fatalf("mirrored API Generated annotation label_format = %d entries, want 3", len(policy.LabelFormat))
	}
	verifyFigmaAPIGeneratedAnnotationContentRule(t, policy, docText)
	verifyMirroredFigmaAPIGeneratedRetiredCategories(t, policy.RetiredCategories, docText)
	verifyFigmaAPIGeneratedLiveInspection(t, policy.LiveInspection, docText)
}

func verifyFigmaAPIGeneratedAnnotationContentRule(t *testing.T, policy figmaSourceAPIGeneratedAnnotationContentRule, docText string) {
	t.Helper()
	ruleText := strings.Join(policy.LabelFormat, "\n") + "\n" + policy.Rule
	for _, needle := range []string{"riido.*", "v2.", "source coverage entry", "종류", "Query", "Mutation", "SSE Stream", "배경", "text/event-stream", "non-stream GET", "non-GET"} {
		if !strings.Contains(ruleText, needle) {
			t.Fatalf("mirrored API Generated annotation content policy must mention %q: %+v", needle, policy)
		}
		if !strings.Contains(docText, figmaAPIGeneratedPolicyDocNeedle(needle)) {
			t.Fatalf("projection doc must mention mirrored API Generated annotation content policy %q", needle)
		}
	}
	if !strings.Contains(policy.Rule, "must not become a second API SSOT") {
		t.Fatalf("mirrored API Generated annotation content policy must prevent second SSOT drift: %q", policy.Rule)
	}
}

func figmaAPIGeneratedPolicyDocNeedle(needle string) string {
	switch needle {
	case "non-stream GET":
		return "non-stream `GET`"
	case "non-GET":
		return "non-`GET`"
	default:
		return needle
	}
}
