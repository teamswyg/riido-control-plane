package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func validTestManifest() manifest {
	return manifest{
		SchemaVersion:    manifestSchema,
		ID:               expectedID,
		Title:            "Context Map",
		RiidoTask:        expectedTask,
		GeneratedDoc:     "docs/20-domain/context-map.md",
		Workflow:         ".github/workflows/context-map.yml",
		EvidenceArtifact: "context-map-evidence",
		OwnedContexts:    testOwnedContexts(),
		ImportedContexts: testImportedContexts(),
		ExternalContexts: testExternalContexts(),
		SSOTLinks:        []link{{Name: "SaaS", Path: "saas-control-plane.md"}},
		SourceChecks:     []sourceCheck{{Name: "anchor", File: "anchors/context.txt", Contains: []string{"context anchor"}}},
		Loop:             evidenceLoop{"o", "h", "x", "e", "r"},
	}
}

func writeTestRepo(t *testing.T, m manifest, freshDoc bool) string {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module fixture\n")
	for _, item := range m.OwnedContexts {
		for _, path := range item.OwnerPaths {
			writeFixtureFile(t, root, path+"/.keep", "owned")
		}
	}
	for _, item := range m.SSOTLinks {
		writeFixtureFile(t, root, "docs/20-domain/"+item.Path, item.Name)
	}
	for _, item := range m.SourceChecks {
		writeFixtureFile(t, root, item.File, item.Contains[0])
	}
	if freshDoc {
		writeFixtureFile(t, root, m.GeneratedDoc, renderDoc(m))
	}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, root, defaultManifest, string(body))
	return root
}

func writeFixtureFile(t *testing.T, root, path, body string) {
	t.Helper()
	target := resolve(root, path)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
