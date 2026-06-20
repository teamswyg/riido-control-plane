package main

import "testing"

func verifyFigmaAPIGeneratedInventoryGeneratedClients(t *testing.T, group figmaSourceAPIGeneratedAnnotationGroup, v2Path string, surface figmaGeneratedClientSurface) {
	t.Helper()
	verifyFacadeGeneratedClientComments(t, "mirrored inventory path "+group.CanonicalGeneratedPath, group.CanonicalGeneratedPath, group.FigmaGeneratedPath, surface)
	verifyFacadeGeneratedClientComments(t, "mirrored inventory path "+group.CanonicalGeneratedPath, v2Path, "riido."+v2Path, surface)
}
