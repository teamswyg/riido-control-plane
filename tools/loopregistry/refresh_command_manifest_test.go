package main

import "testing"

func TestLoopRegistryRefreshCommandsSelectEveryManifestLoopAfterExpiry(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/loop-registry-evidence.json"
	err = run(options{Repo: "../..", Manifest: defaultManifest, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := loadLoopRegistryEvidence(out)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-07-04T00:00:00Z")
	commands, err := selectExpiredRefreshCommands(evidence)
	if err != nil {
		t.Fatal(err)
	}
	if commands.SelectedLoopCount != len(m.Loops) {
		t.Fatalf("selected loops = %d, want %d", commands.SelectedLoopCount, len(m.Loops))
	}
	if !selectedLoopID(commands, "open_decision_queue") {
		t.Fatalf("weekly loop missing from refresh commands: %+v", commands.SelectedLoops)
	}
}

func selectedLoopID(commands refreshCommandEvidence, id string) bool {
	for _, loop := range commands.SelectedLoops {
		if loop.LoopID == id {
			return true
		}
	}
	return false
}
