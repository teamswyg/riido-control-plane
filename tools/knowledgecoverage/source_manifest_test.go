package main

import (
	"strings"
	"testing"
)

func TestSourceManifestRequiresStrictEvidence(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "policy.riido.json", sourceManifestFixture())
	mustWrite(t, root, ".github/workflows/policy.yml", sourceWorkflowFixture("warn"))
	source := sourceSSOT{
		Path:             "policy.riido.json",
		EvidenceTool:     "tools/policycheck",
		Workflow:         ".github/workflows/policy.yml",
		EvidenceArtifact: "policy-evidence",
	}
	problems := validateSourceManifest(root, source)
	if len(problems) == 0 || !strings.Contains(problems[0], "strict artifact") {
		t.Fatalf("problems = %#v", problems)
	}
	mustWrite(t, root, ".github/workflows/policy.yml", sourceWorkflowFixture("error"))
	if problems := validateSourceManifest(root, source); len(problems) != 0 {
		t.Fatalf("strict source manifest failed: %#v", problems)
	}
}

func TestSourceManifestRequiresExecutableMetadata(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "policy.riido.json", "{}")
	mustWrite(t, root, ".github/workflows/policy.yml", sourceWorkflowFixture("error"))
	source := sourceSSOT{
		Path:             "policy.riido.json",
		EvidenceTool:     "tools/policycheck",
		Workflow:         ".github/workflows/policy.yml",
		EvidenceArtifact: "policy-evidence",
	}
	problems := validateSourceManifest(root, source)
	if len(problems) == 0 || !strings.Contains(problems[0], "complete evidence loop") {
		t.Fatalf("problems = %#v", problems)
	}
}

func sourceManifestFixture() string {
	return `{
  "schema_version": "test-policy.v1",
  "id": "test-policy",
  "assertions": ["policy is executable"],
  "loop": {
    "observation": "test",
    "hypothesis": "test",
    "execute": "test",
    "evaluate": "test",
    "retrospective": "test"
  }
}`
}

func sourceWorkflowFixture(mode string) string {
	return "" +
		"steps:\n" +
		"  - run: go run ./tools/policycheck -contract policy.riido.json -evidence-out out/policy.json\n" +
		"  - uses: actions/upload-artifact@v4\n" +
		"    with:\n" +
		"      name: policy-evidence\n" +
		"      path: out/policy.json\n" +
		"      if-no-files-found: " + mode + "\n"
}
