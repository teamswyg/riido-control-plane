package deploypolicy

import (
	"path/filepath"
	"testing"
)

func TestRuntimeCDPublicSurfaceScanContract(t *testing.T) {
	parsed := loadRuntimeCDOwnership(t)
	if parsed.PublicSurfaceScan.RiidoTask != "RIID-4836" {
		t.Fatalf("public surface scan task drifted: %#v", parsed.PublicSurfaceScan)
	}
	if parsed.PublicSensitiveSurfaceGuard.RiidoTask != "RIID-4842" ||
		!parsed.PublicSensitiveSurfaceGuard.PublicKeyNamesAreSensitive {
		t.Fatalf("public sensitive surface guard drifted: %#v", parsed.PublicSensitiveSurfaceGuard)
	}
	assertPublicSurfaceForbiddenContent(t, parsed.PublicSurfaceScan)
	assertPublicConfigKeyScope(t, parsed)
	assertWorkflowForbiddenMechanisms(t, parsed.PublicSurfaceScan.WorkflowForbiddenMechanism)
}

func assertPublicSurfaceForbiddenContent(t *testing.T, scan publicSurfaceScanContract) {
	t.Helper()
	for _, repoPath := range scan.ScopePaths {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, forbidden := range scan.ForbiddenLiterals {
			if contains(body, forbidden) {
				t.Fatalf("%s contains forbidden public CD literal %q", repoPath, forbidden)
			}
		}
		assertNoForbiddenRegex(t, repoPath, body, scan.ForbiddenRegexes)
	}
}

func assertWorkflowForbiddenMechanisms(t *testing.T, forbidden []string) {
	t.Helper()
	for _, workflowPath := range liveWorkflowPaths() {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(workflowPath)))
		for _, phrase := range forbidden {
			if contains(body, phrase) {
				t.Fatalf("%s contains forbidden public CD handoff mechanism %q", workflowPath, phrase)
			}
		}
	}
}
