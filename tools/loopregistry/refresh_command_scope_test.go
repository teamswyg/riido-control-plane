package main

import "testing"

func TestLoopRegistryExpiredRefreshCommandsPreserveScope(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := selectExpiredRefreshCommands(evidence{
		GeneratedAt: "2026-06-23T00:00:00Z",
		ExpiresAt:   "2026-06-24T00:00:00Z",
		RefreshPlans: []refreshPlan{{
			LoopID:            "expired",
			EvidenceExpiresAt: "2026-06-24T00:00:00Z",
			NextCommands: []refreshPlanCommand{{
				Kind:             "target_verifier",
				Command:          "go test ./tools/example",
				ClaimIDs:         []string{"claim-a"},
				EvidenceChainIDs: []string{"chain-a"},
			}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	command := got.Commands[0]
	if command.ClaimIDs[0] != "claim-a" ||
		command.EvidenceChainIDs[0] != "chain-a" {
		t.Fatalf("selected command scope = %+v", command)
	}
}
