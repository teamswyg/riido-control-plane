package main

import (
	"strings"
	"testing"
)

func verifyHeadlessFileKeyLimitation(t *testing.T, ctx figmaSupportingLimitationContext) {
	t.Helper()
	projection, source := requireFigmaSupportingLimitation(t, ctx, "figma-headless-file-key-placeholder.v1")
	if !strings.Contains(source.Tool, "use_figma") ||
		!strings.Contains(source.Tool, "figma.fileKey") ||
		!strings.Contains(source.ObservedResult, "figma.fileKey=headless") ||
		!strings.Contains(source.ObservedResult, "annotation categories") {
		t.Fatalf("source supporting limit must preserve headless file-key evidence: %+v", source)
	}
	for _, result := range []string{"MUOd9lctoEHASUStN3vUuK", "v.1.22 AI Agent"} {
		if !hasString(source.AuthoritativeResult, result) {
			t.Fatalf("source headless file-key authoritative_result must contain %q: %+v", result, source.AuthoritativeResult)
		}
		if !hasString(projection.RequiredAuthoritativeResults, result) {
			t.Fatalf("headless file-key projection must require authoritative result %q: %+v", result, projection.RequiredAuthoritativeResults)
		}
	}
	if strings.TrimSpace(projection.LocalScope) == "" ||
		!strings.Contains(projection.LocalScope, "figma.fileKey=headless") ||
		!strings.Contains(projection.LocalScope, "source identity") {
		t.Fatalf("headless file-key projection must explain local scope: %+v", projection)
	}
	for _, forbidden := range []string{"overwrite source_contracts_manifest.path", "overwrite mirrored figma.file_key", "replace upstream file identity with headless"} {
		if !hasString(projection.ForbiddenProjectionEffects, forbidden) {
			t.Fatalf("headless file-key projection must forbid %q: %+v", forbidden, projection.ForbiddenProjectionEffects)
		}
	}
	requireDocMentions(t, ctx.docText, "headless file-key limitation", []string{
		"figma-headless-file-key-placeholder.v1",
		"`figma.fileKey=headless`",
		"`MUOd9lctoEHASUStN3vUuK`",
		"`figma.file_key`",
		"`source_contracts_manifest`",
		"headless runtime placeholder",
	})
}
