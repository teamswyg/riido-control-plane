package main

import "strings"

func claimSurfaces(
	claims []claimBinding,
	tests map[string][]string,
	chains map[string][]string,
) []claimSurface {
	out := make([]claimSurface, 0, len(claims))
	for _, claim := range claims {
		out = append(out, claimSurfaceFor(claim, tests, chains))
	}
	return out
}

func claimSurfaceFor(
	claim claimBinding,
	tests map[string][]string,
	chains map[string][]string,
) claimSurface {
	surface := claimSurface{
		ID:               claim.ID,
		CodePaths:        []string{},
		TestPaths:        []string{},
		ManifestPaths:    []string{},
		GeneratedDocs:    sortedCopy(claim.GeneratedDoc),
		CoversObserves:   sortedCopy(claim.CoversObserves),
		CoversVerifies:   sortedCopy(claim.CoversVerifies),
		CoversFails:      sortedCopy(claim.CoversFails),
		CoversEvidence:   sortedCopy(claim.CoversEvidence),
		Verifiers:        sortedCopy(claim.Verifiers),
		VerifierCommands: []string{},
		EvidenceChainIDs: sortedCopy(chains[claim.ID]),
	}
	for _, path := range sortedCopy(claim.Files) {
		switch {
		case isClaimTestPath(path):
			surface.TestPaths = append(surface.TestPaths, path)
		case isClaimManifestPath(path):
			surface.ManifestPaths = append(surface.ManifestPaths, path)
		default:
			surface.CodePaths = append(surface.CodePaths, path)
		}
	}
	surface.VerifierCommands = verifierCommandsForClaim(claim, surface.TestPaths, tests)
	return surface
}

func isClaimTestPath(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

func isClaimManifestPath(path string) bool {
	return strings.HasSuffix(path, ".riido.json") || strings.HasPrefix(path, ".github/workflows/")
}
