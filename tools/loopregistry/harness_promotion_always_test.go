package main

import (
	"os"
	"strings"
	"testing"
)

func TestHarnessPromotionStepRequiresAlways(t *testing.T) {
	loop, text, artifact := harnessWorkflowFixtureForTest(t)
	needle := "      - name: Promote harness failures to closed-loop candidates\n        if: always()\n"
	broken := strings.Replace(text, needle, "      - name: Promote harness failures to closed-loop candidates\n", 1)
	if broken == text {
		t.Fatalf("promotion step fixture %s did not contain if: always()", loop.RefreshWorkflow)
	}
	if harnessWorkflowPromotesCandidates(broken, artifact) {
		t.Fatal("expected promotion step without if: always() to fail")
	}
}

func TestHarnessCandidateArtifactUploadRequiresAlways(t *testing.T) {
	loop, text, artifact := harnessWorkflowFixtureForTest(t)
	needle := "      - uses: actions/upload-artifact@v7\n        if: always()\n        with:\n          name: " + artifact
	broken := strings.Replace(text, needle, "      - uses: actions/upload-artifact@v7\n        with:\n          name: "+artifact, 1)
	if broken == text {
		t.Fatalf("candidate artifact upload fixture %s did not contain if: always()", loop.RefreshWorkflow)
	}
	if harnessWorkflowPromotesCandidates(broken, artifact) {
		t.Fatal("expected candidate upload step without if: always() to fail")
	}
}

func TestHarnessPromotionWorkflowPublishesEvidenceArtifact(t *testing.T) {
	text := workflowTextForTest(t, ".github/workflows/harness-promotion.yml")
	if !refreshWorkflowDeclaresLoopID(text, "closed_loop_candidate") {
		t.Fatal("harness-promotion workflow must declare closed_loop_candidate loop identity")
	}
	if !workflowUploadsStrictArtifact(text, "harness-promotion-evidence") {
		t.Fatal("harness-promotion workflow must publish strict harness-promotion-evidence artifact")
	}
}

func harnessWorkflowFixtureForTest(t *testing.T) (loopRecord, string, string) {
	t.Helper()
	m, _ := loadLoopRegistryForTest(t)
	loop := m.Loops[loopIndex(t, m, "provider_acceptance_harness")]
	artifact, err := harnessCandidateArtifact(loop)
	if err != nil {
		t.Fatalf("harness candidate artifact: %v", err)
	}
	return loop, harnessWorkflowTextForTest(t, loop), artifact
}

func harnessWorkflowTextForTest(t *testing.T, loop loopRecord) string {
	t.Helper()
	return workflowTextForTest(t, loop.RefreshWorkflow)
}

func workflowTextForTest(t *testing.T, workflow string) string {
	t.Helper()
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(repoPath(root, workflow))
	if err != nil {
		t.Fatalf("read harness workflow: %v", err)
	}
	return string(data)
}
