package main

import (
	"strings"
	"testing"
)

func verifyOnboardingTimeoutLimitation(t *testing.T, ctx figmaSupportingLimitationContext) {
	t.Helper()
	projection, source := requireFigmaSupportingLimitation(t, ctx, "figma-onboarding-page-load-timeout.v1")
	if !strings.Contains(source.Tool, "42:3014") ||
		!strings.Contains(source.ObservedResult, "time out after 120s") ||
		!strings.Contains(source.ObservedResult, "Wireframe - 온보딩") ||
		!strings.Contains(source.ObservedResult, "236:33845") ||
		!strings.Contains(source.ObservedResult, "236:33847") {
		t.Fatalf("source supporting limit must preserve onboarding load timeout evidence: %+v", source)
	}
	verifyOnboardingAuthoritativeResults(t, projection, source)
	verifyOnboardingLocalScope(t, projection)
	verifyOnboardingForbiddenEffects(t, projection)
	requireDocMentions(t, ctx.docText, "onboarding page load timeout", []string{
		"figma-onboarding-page-load-timeout.v1",
		"after 120s",
		"`Wireframe - 온보딩`",
		"`236:33845`",
		"`236:33847`",
		"six onboarding `riido.*` `API",
		"`non_ui_top_level_inventory`",
		"mark onboarding generated paths unresolved",
		"`child_count=84`",
		"`known_inventory_count=83`",
	})
}

func verifyOnboardingAuthoritativeResults(t *testing.T, projection figmaProjectionSupportingToolLimitation, source figmaSourceSupportingToolLimitation) {
	t.Helper()
	for _, result := range []string{"42:3014", "child_count=84", "known_inventory_count=83", "unresolved_extra_top_level_node=1", "non_ui_top_level_inventory", "236:33845", "236:33847", "onboarding_api_generated_annotations=6"} {
		if !hasString(source.AuthoritativeResult, result) {
			t.Fatalf("source onboarding timeout authoritative_result must contain %q: %+v", result, source.AuthoritativeResult)
		}
	}
	for _, pageID := range projection.RequiredAuthoritativePages {
		if !hasString(source.AuthoritativeResult, pageID) {
			t.Fatalf("onboarding timeout projection page %q is absent from source authoritative_result: %+v", pageID, source.AuthoritativeResult)
		}
	}
	for _, result := range []string{"child_count=84", "known_inventory_count=83", "unresolved_extra_top_level_node=1", "onboarding_api_generated_annotations=6"} {
		if !hasString(projection.RequiredAuthoritativeResults, result) {
			t.Fatalf("onboarding timeout projection must require authoritative result %q: %+v", result, projection.RequiredAuthoritativeResults)
		}
	}
}
