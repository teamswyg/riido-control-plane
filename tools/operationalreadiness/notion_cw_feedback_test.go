package main

import "testing"

func TestOperationalReadinessBindsCWFeedbackInventory(t *testing.T) {
	m := loadManifestForTest(t)
	check := findReadinessCheck(t, m, "staging_client_p0_visual_retest")
	want := []string{
		"cw_aia_001_existing_user_feature_entry",
		"cw_aia_002_existing_user_list_redirect",
		"cw_aia_003_default_agent_selection_confirmed",
		"cw_aia_004_description_clamp_open",
		"cw_aia_005_claude_logo_contrast_open",
		"cw_aia_006_low_height_cta_overlap_confirmed",
		"cw_aia_007_claude_asset_open",
		"cw_aia_008_screen_composition_open",
		"cw_aia_009_codex_onboarding_runtime_confirmed",
		"cw_aia_010_long_nickname_ellipsis_open",
		"cw_aia_011_right_pane_ai_logo_open",
	}
	got := map[string]bool{}
	for _, measurement := range check.Measurements {
		if measurement.Kind == "notion" {
			got[measurement.ID] = true
		}
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("missing CW feedback measurement %s", id)
		}
	}
}

func TestOperationalReadinessBindsCWFeedbackNotionCycle(t *testing.T) {
	m := loadManifestForTest(t)
	for _, cycle := range m.NotionOpenLoop.Cycles {
		if cycle.ID == "notion_p0_cw_feedback_inventory" &&
			cycle.BackfilledCheck == "staging_client_p0_visual_retest" {
			return
		}
	}
	t.Fatal("missing CW feedback Notion backfill cycle")
}
