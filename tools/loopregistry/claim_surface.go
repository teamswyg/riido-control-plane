package main

import "strings"

func claimSurfaces(claims []claimBinding) []claimSurface {
	out := make([]claimSurface, 0, len(claims))
	for _, claim := range claims {
		out = append(out, claimSurfaceFor(claim))
	}
	return out
}

func claimSurfaceFor(claim claimBinding) claimSurface {
	surface := claimSurface{
		ID:            claim.ID,
		CodePaths:     []string{},
		TestPaths:     []string{},
		ManifestPaths: []string{},
		GeneratedDocs: sortedCopy(claim.GeneratedDoc),
		Verifiers:     sortedCopy(claim.Verifiers),
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
	return surface
}

func isClaimTestPath(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

func isClaimManifestPath(path string) bool {
	return strings.HasSuffix(path, ".riido.json") || strings.HasPrefix(path, ".github/workflows/")
}
