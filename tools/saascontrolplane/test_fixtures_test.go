package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/requirements"
)

func validTestManifest() manifest {
	workflows := testWorkflows()
	return manifest{
		SchemaVersion:    requirements.ManifestSchema,
		ID:               requirements.ExpectedID,
		Title:            "SaaS Control Plane",
		RiidoTasks:       []string{"RIID-4712"},
		GeneratedDoc:     domainDocPath,
		Workflow:         ".github/workflows/saas-control-plane.yml",
		EvidenceArtifact: "saas-control-plane-evidence",
		OwnerPackage:     "internal/riidoaiserver",
		SharedContracts:  requirements.RequiredSharedContracts,
		FocusedWorkflows: workflows,
		Boundaries:       testBoundaries(workflows),
		RequiredPhrases:  []phrase{{File: domainDocPath, Contains: "owner package"}},
		Loop:             evidenceLoop{"o", "h", "x", "e", "r"},
	}
}

func writeTestRepo(t *testing.T, m manifest, freshDoc bool) string {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, requirements.DefaultManifest, mustJSON(t, m))
	for _, workflow := range m.FocusedWorkflows {
		body := domainDocPath + "\ngo run ./tools/dependencyallowlist\n" + testArtifactText(m)
		writeFixtureFile(t, root, workflow, body)
	}
	for _, item := range m.Boundaries {
		for _, check := range item.SourceChecks {
			writeFixtureFile(t, root, check.File, check.Contains[0])
		}
	}
	if freshDoc {
		writeFixtureFile(t, root, m.GeneratedDoc, renderDoc(m))
	}
	return root
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func writeFixtureFile(t *testing.T, root, path, body string) {
	t.Helper()
	target := pathutilRepoPath(root, path)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func pathutilRepoPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
