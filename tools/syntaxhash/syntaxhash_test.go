package main

import "testing"

func TestSyntaxHashGraphUsesGoldenLockedContextMap(t *testing.T) {
	out := t.TempDir() + "/syntax-hash.json"
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	var graph syntaxGraph
	if err := readJSON(out, &graph); err != nil {
		t.Fatal(err)
	}
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
	if graph.Score.WeightedScore != graph.Score.EfficiencyScore+graph.Score.CompressionScore ||
		graph.Score.AnalysisReduction == 0 || graph.Score.ConstraintGate == "" ||
		graph.Score.Formula == "" {
		t.Fatalf("incomplete score evidence: %+v", graph.Score)
	}
	if graph.Score.CollisionCount != 0 || graph.Score.MissingRelocations != 0 ||
		graph.Score.RelocationMappings != graph.Score.TrackedFiles {
		t.Fatalf("incomplete safety evidence: %+v", graph.Score)
	}
	if graph.Score.MissingGoldens != 0 || graph.Score.GoldenCommands != len(graph.Targets) {
		t.Fatalf("incomplete golden evidence: %+v", graph.Score)
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
	m.GeneratedDoc = t.TempDir() + "/syntax-hash.md"
	m.Targets = m.Targets[:1]
	m.Constraints.MinRepositoryCoverageBasisPoint = 1
	manifestPath := t.TempDir() + "/manifest.json"
	if err := writeJSON(manifestPath, m); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/syntax-hash.json"
	if err := run(options{Repo: "../..", Manifest: manifestPath, EvidenceOut: out}); err != nil {
		t.Fatal(err)
	}
	var graph syntaxGraph
	if err := readJSON(out, &graph); err != nil {
		t.Fatal(err)
	}
	if got := graph.Targets[0].PackageHash; got != "a11308340a6110b8f76064273920d7d2d4fb057cb654e525997679659fd1f622" {
		t.Fatalf("contextmap syntax package hash drifted: %s", got)
	}
}

func TestSyntaxHashGeneratedDocMatches(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
