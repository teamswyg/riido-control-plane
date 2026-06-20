package main

import (
	"strings"
	"testing"
)

func verifyOnboardingLocalScope(t *testing.T, projection figmaProjectionSupportingToolLimitation) {
	t.Helper()
	if strings.TrimSpace(projection.LocalScope) == "" ||
		!strings.Contains(projection.LocalScope, "must not treat") ||
		!strings.Contains(projection.LocalScope, "42:3014") ||
		!strings.Contains(projection.LocalScope, "child_count=84") ||
		!strings.Contains(projection.LocalScope, "known_inventory_count=83") {
		t.Fatalf("onboarding timeout projection must explain local scope: %+v", projection)
	}
}

func verifyOnboardingForbiddenEffects(t *testing.T, projection figmaProjectionSupportingToolLimitation) {
	t.Helper()
	for _, forbidden := range []string{"remove expected_pages", "remove non_ui_top_level_inventory", "remove onboarding generated paths", "mark onboarding generated paths unresolved"} {
		if !hasString(projection.ForbiddenProjectionEffects, forbidden) {
			t.Fatalf("onboarding timeout projection must forbid %q: %+v", forbidden, projection.ForbiddenProjectionEffects)
		}
	}
}
