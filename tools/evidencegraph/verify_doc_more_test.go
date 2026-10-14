package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyManifestRejectsRequiredFields(t *testing.T) {
	t.Parallel()
	base := manifest{
		SchemaVersion: manifestSchema, ID: "id", Title: "title", GeneratedDoc: "doc.md",
		Workflow: ".github/workflows/evidence-graph.yml", Evidence: "artifact",
		EvidenceTool: "tools/evidencegraph", LoopRegistry: "docs/30-architecture/loop-registry.riido.json",
		Assertions: []string{"a"}, Chains: []chain{testChain("chain", "observation")},
		Loop: loopRecord{
			Observation: "o", Hypothesis: "h", Execute: "x", Evaluate: "e", Retrospective: "r",
		},
	}
	cases := []struct {
		name string
		edit func(*manifest)
		want string
	}{
		{"schema", func(m *manifest) { m.SchemaVersion = "bad" }, "schema_version"},
		{"identity", func(m *manifest) { m.ID = "" }, "id, title"},
		{"workflow", func(m *manifest) { m.EvidenceTool = "other" }, "workflow"},
		{"registry", func(m *manifest) { m.LoopRegistry = "" }, "loop_registry"},
		{"content", func(m *manifest) { m.Chains = nil }, "assertions"},
		{"loop execute", func(m *manifest) { m.Loop.Execute = "" }, "observation"},
		{"loop retro", func(m *manifest) { m.Loop.Retrospective = "" }, "retrospective"},
	}
	for _, tc := range cases {
		got := base
		tc.edit(&got)
		if err := verifyManifest(got); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s error = %v, want %q", tc.name, err, tc.want)
		}
	}
}

func TestMaybeDocWritesChecksAndRejectsDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	m := manifest{GeneratedDoc: "docs/generated.md"}
	if err := maybeDoc(root, m, "doc", true, true); err != nil {
		t.Fatalf("maybeDoc write/check: %v", err)
	}
	if err := maybeDoc(root, m, "other", false, true); err == nil {
		t.Fatalf("maybeDoc should reject stale doc")
	}
	m.GeneratedDoc = filepath.Join("missing", "doc.md")
	if err := maybeDoc(root, m, "doc", false, true); err == nil {
		t.Fatalf("maybeDoc should reject missing generated doc")
	}
}
