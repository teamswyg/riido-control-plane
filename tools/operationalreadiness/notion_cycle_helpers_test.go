package main

import "testing"

func notionCycleHasEvidenceRef(cycle notionCycle, want evidenceRef) bool {
	for _, got := range cycle.EvidenceRefs {
		if got == want {
			return true
		}
	}
	return false
}

func findNotionCycle(t *testing.T, m manifest, id string) notionCycle {
	t.Helper()
	for _, cycle := range m.NotionOpenLoop.Cycles {
		if cycle.ID == id {
			return cycle
		}
	}
	t.Fatalf("missing Notion cycle %s", id)
	return notionCycle{}
}
