package deploypolicy

import "testing"

func assertRuntimeCDSensitiveSurface(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	guard := p.PublicSensitiveSurfaceGuard
	if guard.RiidoTask != "RIID-4842" {
		t.Fatalf("public sensitive surface guard work unit drifted: %#v", guard)
	}
	if guard.CanonicalOwner != "riido-control-plane" || guard.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public sensitive surface guard ownership drifted: %#v", guard)
	}
	if !guard.PublicKeyNamesAreSensitive {
		t.Fatal("public sensitive surface guard must treat public key names as sensitive")
	}
	requireContains(t, guard.Rule, "sensitivity budget")
	assertSensitiveAllowedInfo(t, guard)
	assertSensitiveScopePaths(t, guard)
	assertSensitiveForbiddenInfo(t, guard)
	assertNonCDRuntimeKeys(t, guard.AllowedPublicNonCDRuntime)
}

func assertSensitiveAllowedInfo(t *testing.T, guard publicSensitiveSurfaceGuard) {
	t.Helper()
	requireSliceContains(t, guard.AllowedPublicInformation, "current stable RIIDO_AI_SERVER_* key names listed in public_config_key_minimization only in canonical machine-readable and workflow paths")
	requireSliceContains(t, guard.AllowedPublicInformation, "explicit non-CD runtime key exceptions listed in this guard")
	requireSliceContains(t, guard.InfraMustKnow, "CD execution remains owned by riido-control-plane")
	requireSliceContains(t, guard.InfraMustKnow, "infra consumes the stable key categories and source names only")
	requireSliceContains(t, guard.InfraMustKnow, "human-readable public docs link to the manifest instead of repeating exact deploy/smoke key lists")
}

func assertSensitiveScopePaths(t *testing.T, guard publicSensitiveSurfaceGuard) {
	t.Helper()
	requireSliceContains(t, guard.KeyNameScopePaths, "README.md")
	requireSliceContains(t, guard.KeyNameScopePaths, "docs/30-architecture/runtime-cd-ownership.md")
	requireSliceContains(t, guard.KeyNameScopePaths, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, guard.CanonicalCDKeyListPaths, "docs/30-architecture/runtime-cd-ownership.riido.json")
	requireSliceContains(t, guard.CanonicalCDKeyListPaths, ".github/workflows/deploy-ai-agent-testnet.yml")
	if stringSliceContains(guard.CanonicalCDKeyListPaths, "docs/30-architecture/runtime-cd-ownership.md") ||
		stringSliceContains(guard.CanonicalCDKeyListPaths, "docs/30-architecture/runtime-deployment-boundary.md") {
		t.Fatalf("human-readable docs must not be canonical exact CD key-list paths: %#v", guard.CanonicalCDKeyListPaths)
	}
}

func assertSensitiveForbiddenInfo(t *testing.T, guard publicSensitiveSurfaceGuard) {
	t.Helper()
	requireSliceContains(t, guard.ForbiddenPublicInformation, "new RIIDO_AI_SERVER_* key names that are not listed in public_config_key_minimization")
	requireSliceContains(t, guard.ForbiddenPublicInformation, "exact deploy/smoke key-name lists in human-readable public docs outside canonical machine-readable and workflow paths")
	requireSliceContains(t, guard.ForbiddenPublicInformation, "example values for allowed public key names")
}
