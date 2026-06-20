package main

import (
	"strings"
	"testing"
)

func verifyMetadataPageListLimitation(t *testing.T, ctx figmaSupportingLimitationContext) {
	t.Helper()
	projection, source := requireFigmaSupportingLimitation(t, ctx, "figma-metadata-page-list-underreports-pages.v1")
	if !strings.Contains(source.Tool, "get_metadata") || !strings.Contains(source.ObservedResult, "only page 129:5215 UI") {
		t.Fatalf("source supporting limit must preserve no-nodeId metadata under-report evidence: %+v", source)
	}
	if !hasString(ctx.stabilizedBy, "teamswyg/riido-contracts#52") {
		t.Fatalf("source_contracts_manifest.stabilized_by must include contracts #52 for mirrored metadata limitation: %+v", ctx.stabilizedBy)
	}
	if strings.TrimSpace(projection.LocalScope) == "" || !strings.Contains(projection.LocalScope, "must not collapse") {
		t.Fatalf("projection supporting limit must explain local scope: %+v", projection)
	}
	verifyMetadataRequiredPages(t, projection, source)
	for _, forbidden := range []string{"remove expected_pages", "remove non_ui_top_level_inventory", "remove legacy_non_ui_absorptions"} {
		if !hasString(projection.ForbiddenProjectionEffects, forbidden) {
			t.Fatalf("projection supporting limit must forbid %q: %+v", forbidden, projection.ForbiddenProjectionEffects)
		}
	}
	requireDocMentions(t, ctx.docText, "mirrored supporting tool limitation", []string{
		"figma-metadata-page-list-underreports-pages.v1",
		"get_metadata",
		"teamswyg/riido-contracts#52",
		"`129:5215`",
		"`42:3014`",
		"`0:1`",
		"`expected_pages`",
		"`non_ui_top_level_inventory`",
		"`legacy_non_ui_absorptions`",
	})
}

func verifyMetadataRequiredPages(t *testing.T, projection figmaProjectionSupportingToolLimitation, source figmaSourceSupportingToolLimitation) {
	t.Helper()
	requiredPages := map[string]bool{"129:5215": false, "42:3014": false, "0:1": false}
	for _, pageID := range projection.RequiredAuthoritativePages {
		if _, ok := requiredPages[pageID]; ok {
			requiredPages[pageID] = true
		}
		if !hasString(source.AuthoritativeResult, pageID) {
			t.Fatalf("projection supporting limit page %q is absent from source authoritative_result: %+v", pageID, source.AuthoritativeResult)
		}
	}
	for pageID, seen := range requiredPages {
		if !seen {
			t.Fatalf("projection supporting limit is missing authoritative page %s", pageID)
		}
	}
}
