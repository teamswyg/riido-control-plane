package main

import "testing"

func TestCandidateDecisionRefreshCommandsUseNextCommand(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	result, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), out)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	got := newRefreshCommandEvidence(result)
	if got.Status != "refresh_required" || got.CommandCount != 1 {
		t.Fatalf("unexpected command evidence: %+v", got)
	}
	if got.Commands[0].Command != result.DecisionArtifacts[0].NextCommand {
		t.Fatalf("command mismatch: %+v", got.Commands[0])
	}
	if got.Commands[0].LoopID != result.DecisionArtifacts[0].NextLoop {
		t.Fatalf("loop mismatch: %+v", got.Commands[0])
	}
	if got.Commands[0].Kind != "target_verifier" {
		t.Fatalf("kind = %q", got.Commands[0].Kind)
	}
}

func TestCandidateDecisionCommandsOutWritesRefreshEvidence(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	commandsOut := t.TempDir() + "/commands.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CandidateIn: out, CommandsOut: commandsOut}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got refreshCommandEvidence
	if err := readJSON(commandsOut, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != refreshCommandSchema || got.CommandCount != 1 {
		t.Fatalf("commands evidence = %+v", got)
	}
}
