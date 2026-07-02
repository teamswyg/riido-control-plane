package main

import "testing"

func TestOperationalReadinessMarksServerCompleteP0AsTransferRequested(t *testing.T) {
	m := loadManifestForTest(t)
	want := []string{
		"notion_p0_queued_by_busy_agent",
		"notion_p0_terminal_stop_refresh",
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

func TestOperationalReadinessBindsQueuedStateOnlyEvidence(t *testing.T) {
	m := loadManifestForTest(t)
	cycle := findNotionCycle(t, m, "notion_p0_queued_by_busy_agent")
	want := []evidenceRef{
		{Kind: "test", Path: "internal/riidoaiserver/ai_agent_client_thread_history_queued_status_test.go"},
		{Kind: "external", Path: "github:https://github.com/teamswyg/riido-control-plane/pull/745"},
		{Kind: "external", Path: "github:https://github.com/teamswyg/riido-control-plane/actions/runs/28565727298"},
		{Kind: "external", Path: "notion-comment:39120241-cf7f-81f7-b37a-001dba758d28"},
	}
	for _, ref := range want {
		if !notionCycleHasEvidenceRef(cycle, ref) {
			t.Fatalf("queued P0 cycle missing evidence ref %+v", ref)
		}
	}
}

func TestOperationalReadinessBindsTerminalStopLiveEvidence(t *testing.T) {
	m := loadManifestForTest(t)
	cycle := findNotionCycle(t, m, "notion_p0_terminal_stop_refresh")
	want := []evidenceRef{
		{Kind: "test", Path: "internal/riidoaiserver/ai_agent_client_http_v3_terminal_stream_test.go"},
		{Kind: "test", Path: "internal/riidoaiserver/ai_agent_client_http_v3_stop_refresh_test.go"},
		{Kind: "api_snapshot", Path: "docs/30-architecture/evidence/staging-api-terminal-stop-refresh-real-mutation-2026-07-02.json"},
		{Kind: "external", Path: "notion-comment:39120241-cf7f-81d3-b312-001dec529267"},
	}
	for _, ref := range want {
		if !notionCycleHasEvidenceRef(cycle, ref) {
			t.Fatalf("terminal stop P0 cycle missing evidence ref %+v", ref)
		}
	}
}

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
