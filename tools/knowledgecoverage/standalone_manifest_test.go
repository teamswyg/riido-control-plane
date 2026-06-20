package main

import (
	"strings"
	"testing"
)

func TestStandaloneManifestRequiresStrictEvidenceUpload(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "docs/risk.riido.json", standaloneManifestFixture())
	mustWrite(t, root, "docs/risk.md", "# Risk\n")
	mustWrite(t, root, "tools/risk/main.go", "package main\n")
	mustWrite(t, root, ".github/workflows/risk.yml", standaloneWorkflowFixture("warn"))
	m := manifest{Standalone: []standalone{standaloneFixture()}}
	problems := validateStandaloneManifests(root, m)
	if len(problems) == 0 || !strings.Contains(strings.Join(problems, "\n"), "strict artifact step") {
		t.Fatalf("problems = %#v", problems)
	}
}

func standaloneFixture() standalone {
	return standalone{
		Path:             "docs/risk.riido.json",
		EvidenceTool:     "tools/risk",
		Workflow:         ".github/workflows/risk.yml",
		EvidenceArtifact: "risk-evidence",
		HumanDoc:         "docs/risk.md",
	}
}

func standaloneManifestFixture() string {
	return `{"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`
}

func standaloneWorkflowFixture(missingMode string) string {
	return "steps:\n" +
		"  - run: go run ./tools/risk -check-doc -evidence-out out/risk.json\n" +
		"  - uses: actions/upload-artifact@v4\n" +
		"    with:\n" +
		"      name: risk-evidence\n" +
		"      path: out/risk.json\n" +
		"      if-no-files-found: " + missingMode + "\n"
}
