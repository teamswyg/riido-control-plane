package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readGraph(t *testing.T, path string) syntaxGraph {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var graph syntaxGraph
	if err := json.Unmarshal(body, &graph); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestSyntaxHashGraphUsesGoldenLockedContextMap(t *testing.T) {
	out := filepath.Join(t.TempDir(), "syntax-hash.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	graph := readGraph(t, out)
	if graph.Status != "verified" || len(graph.Targets) < 2 {
		t.Fatalf("unexpected graph: %+v", graph)
	}
	target := graph.Targets[0]
	if target.Coverage != 100 || target.PackageHash == "" || target.SemanticHash == "" {
		t.Fatalf("incomplete target graph: %+v", target)
	}
	if graph.Repository.GoFiles != graph.Repository.TrackedFiles ||
		graph.Repository.UntrackedFiles != 0 || len(graph.Repository.UntrackedSample) != 0 {
		t.Fatalf("expected repository syntax coverage to be complete: %+v", graph.Repository)
	}
	if graph.Constraints.MinRepositoryCoverageBasisPoint != 10000 {
		t.Fatalf("repository coverage floor = %+v", graph.Constraints)
	}
	if len(target.Relocations) != target.TrackedFiles {
		t.Fatalf("missing relocation evidence: %+v", target.Relocations)
	}
}

func TestSyntaxHashToolBehaviorGolden(t *testing.T) {
	var m manifest
	if err := readJSON(repoPath("../..", defaultManifest), &m); err != nil {
		t.Fatal(err)
	}
	m.GeneratedDoc = filepath.Join(t.TempDir(), "syntax-hash.md")
	m.Targets = m.Targets[:1]
	m.Constraints.MinRepositoryCoverageBasisPoint = 1
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeJSON(manifestPath, m); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "syntax-hash.json")
	if err := run(options{Repo: "../..", Manifest: manifestPath, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
	graph := readGraph(t, out)
	if got := graph.Targets[0].PackageHash; got != "a11308340a6110b8f76064273920d7d2d4fb057cb654e525997679659fd1f622" {
		t.Fatalf("contextmap syntax package hash drifted: %s", got)
	}
}

func TestSyntaxHashGeneratedDocMatches(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
