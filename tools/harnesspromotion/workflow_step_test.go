package main

import (
	"strings"
	"testing"
)

func TestPromotionWorkflowRequiresAlwaysCandidateStep(t *testing.T) {
	source, text := promotionWorkflowFixture(t)
	broken := strings.Replace(text, "      - name: Promote harness failures to closed-loop candidates\n        if: always()\n", "      - name: Promote harness failures to closed-loop candidates\n", 1)
	if broken == text {
		t.Fatal("fixture did not contain always-running promotion step")
	}
	if err := verifySourceWorkflow(broken, source); err == nil {
		t.Fatal("expected missing always-running promotion step to fail")
	}
}

func TestPromotionWorkflowRequiresAlwaysCandidateUpload(t *testing.T) {
	source, text := promotionWorkflowFixture(t)
	needle := "      - uses: actions/upload-artifact@v7\n        if: always()\n        with:\n          name: " + source.CandidateArtifact
	broken := strings.Replace(text, needle, "      - uses: actions/upload-artifact@v7\n        with:\n          name: "+source.CandidateArtifact, 1)
	if broken == text {
		t.Fatal("fixture did not contain always-running candidate upload")
	}
	if err := verifySourceWorkflow(broken, source); err == nil {
		t.Fatal("expected missing always-running candidate upload to fail")
	}
}

func promotionWorkflowFixture(t *testing.T) (promotionSource, string) {
	t.Helper()
	m := loadPromotionManifestForTest(t)
	source := m.Sources[0]
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	text, err := workflowText(root, source.SourceWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	return source, text
}
