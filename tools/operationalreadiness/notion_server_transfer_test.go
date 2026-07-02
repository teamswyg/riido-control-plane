package main

import "testing"

func TestOperationalReadinessMarksServerCompleteP0AsTransferRequested(t *testing.T) {
	m := loadManifestForTest(t)
	want := []string{
		"notion_p0_queued_by_busy_agent",
		"notion_p0_instruction_intent_dialo" + "gue",
	}
	for _, id := range want {
		cycle := findNotionCycle(t, m, id)
		if cycle.CodexStatus != "[codex][완료][전달요청]" {
			t.Fatalf("%s codex_status = %q", id, cycle.CodexStatus)
		}
		if cycle.Status != "partial" {
			t.Fatalf("%s status = %q, want partial until visual evidence", id, cycle.Status)
		}
		if cycle.RequiredNextArtifact != "staging_client_p0_visual_screenshot_evidence" {
			t.Fatalf("%s next artifact = %q", id, cycle.RequiredNextArtifact)
		}
	}
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
