package main

import (
	"encoding/json"
	"os"
	"testing"
)

const notionDMGWarningUpdateEvidence = "docs/30-architecture/evidence/notion-dmg-stale-warning-update-2026-07-02.json"

func TestOperationalReadinessBindsNotionDMGWarningUpdate(t *testing.T) {
	check := readinessCheckByID(t, "staging_client_p0_visual_retest")
	if !hasMeasurement(check, "notion_dmg_stale_warning_update_2026_07_02") {
		t.Fatal("missing Notion DMG warning update measurement")
	}
	if !hasEvidenceRef(check, notionDMGWarningUpdateEvidence) {
		t.Fatal("missing Notion DMG warning update evidence ref")
	}
	if check.Status != "partial" {
		t.Fatalf("visual readiness must remain partial until teammate install and visual QA close: %s", check.Status)
	}
}

func TestNotionDMGWarningUpdateRecordsStaleRemoval(t *testing.T) {
	body, err := os.ReadFile("../../" + notionDMGWarningUpdateEvidence)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Redacted          bool `json:"redacted"`
		VerifiedPageState struct {
			TopWarningPresent                 bool `json:"top_warning_present"`
			TopWarningMentionsSignedNotarized bool `json:"top_warning_mentions_signed_notarized_dmg"`
			TopWarningMentionsStaleDoNotUse   bool `json:"top_warning_mentions_stale_workaround_do_not_use_first"`
			StaleWorkaroundStillPresent       bool `json:"stale_workaround_block_still_present"`
			StaleWorkaroundCommandsRemoved    bool `json:"stale_workaround_commands_removed"`
			StaleWorkaroundOnlyDoNotUse       bool `json:"stale_workaround_only_mentioned_as_do_not_use"`
		} `json:"verified_page_state"`
		Decision struct {
			Status string `json:"status"`
		} `json:"decision"`
	}
	if err := json.Unmarshal(body, &evidence); err != nil {
		t.Fatal(err)
	}
	if !evidence.Redacted || evidence.Decision.Status != "covered" {
		t.Fatalf("unexpected evidence status: redacted=%v decision=%s", evidence.Redacted, evidence.Decision.Status)
	}
	state := evidence.VerifiedPageState
	if !state.TopWarningPresent || !state.TopWarningMentionsSignedNotarized || !state.TopWarningMentionsStaleDoNotUse {
		t.Fatalf("top warning is not strong enough: %+v", state)
	}
	if state.StaleWorkaroundStillPresent {
		t.Fatal("evidence must not leave the stale workaround block as an active install step")
	}
	if !state.StaleWorkaroundCommandsRemoved || !state.StaleWorkaroundOnlyDoNotUse {
		t.Fatalf("stale workaround command state is not closed: %+v", state)
	}
}
