package main

import "testing"

func TestLoopRegistrySelectsExpiredRefreshCommands(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := selectExpiredRefreshCommands(evidence{
		GeneratedAt: "2026-06-23T00:00:00Z",
		ExpiresAt:   "2026-06-24T00:00:00Z",
		RefreshPlans: []refreshPlan{{
			LoopID:            "expired",
			EvidenceExpiresAt: "2026-06-24T00:00:00Z",
			NextCommands: []refreshPlanCommand{{
				Kind:    "refresh_workflow",
				Command: "gh workflow run loop-registry.yml --ref main",
			}},
		}, {
			LoopID:            "fresh",
			EvidenceExpiresAt: "2026-06-26T00:00:00Z",
			NextCommands:      []refreshPlanCommand{{Kind: "target_verifier", Command: "go test ./..."}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "refresh_required" || got.CommandCount != 1 || got.Commands[0].LoopID != "expired" {
		t.Fatalf("refresh commands = %+v", got)
	}
	if got.GeneratedAt != "2026-06-25T00:00:00Z" || got.ExpiresAt != "2026-06-26T00:00:00Z" {
		t.Fatalf("refresh command freshness = %+v", got)
	}
}

func TestLoopRegistryRefreshCommandModeWritesEvidence(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	dir := t.TempDir()
	in := dir + "/source.json"
	out := dir + "/commands.json"
	if err := writeJSON(in, evidence{
		GeneratedAt: "2026-06-23T00:00:00Z",
		ExpiresAt:   "2026-06-24T00:00:00Z",
		RefreshPlans: []refreshPlan{{
			LoopID:            "expired",
			EvidenceExpiresAt: "2026-06-24T00:00:00Z",
			NextCommands:      []refreshPlanCommand{{Kind: "refresh_workflow", Command: "run"}},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := mainRun([]string{"-refresh-plan-in", in, "-refresh-commands-out", out}); err != nil {
		t.Fatal(err)
	}
	written, err := loadRefreshCommandEvidence(out)
	if err != nil {
		t.Fatal(err)
	}
	if written.CommandCount != 1 || written.SchemaVersion != refreshCommandsSchema || written.ExpiresAt == "" {
		t.Fatalf("written refresh command evidence = %+v", written)
	}
}
