package main

import "testing"

func TestOperationalReadinessBindsStagingDMGGatekeeperEvidence(t *testing.T) {
	m := loadManifestForTest(t)
	cycle := findNotionCycle(t, m, "notion_p0_gatekeeper_staging_dmg")
	if cycle.CodexStatus != "[codex][완료][전달요청]" {
		t.Fatalf("gatekeeper codex_status = %q", cycle.CodexStatus)
	}
	if cycle.Status != "partial" {
		t.Fatalf("gatekeeper status = %q, want partial until install screenshot", cycle.Status)
	}
	want := []evidenceRef{
		{Kind: "external", Path: "cdn:https://cdn.riido.io/releases/latest/staging/Riido-Staging-arm64.dmg"},
		{Kind: "external", Path: "sha256:7a20806e75c35fef95a2f680476563b08b37d31c6bc23ae057e67fa0e36fa65a"},
		{Kind: "external", Path: "spctl:dmg-accepted-source-notarized-developer-id"},
		{Kind: "external", Path: "spctl:app-accepted-source-notarized-developer-id"},
		{Kind: "external", Path: "stapler:dmg-validate-action-worked"},
		{Kind: "external", Path: "stapler:app-validate-action-worked"},
		{Kind: "external", Path: "codesign:developer-id-application-swyg-inc-3pvv8dqpl6"},
		{Kind: "external", Path: "notion-comment:39120241-cf7f-81f2-9d54-001d551d544e"},
	}
	for _, ref := range want {
		if !notionCycleHasEvidenceRef(cycle, ref) {
			t.Fatalf("gatekeeper P0 cycle missing evidence ref %+v", ref)
		}
	}
}
