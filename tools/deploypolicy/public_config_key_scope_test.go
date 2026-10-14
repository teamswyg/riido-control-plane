package deploypolicy

import (
	"path/filepath"
	"strings"
	"testing"
)

func assertPublicConfigKeyScope(t *testing.T, parsed runtimeCDOwnership) {
	t.Helper()
	allowedPublicKeys := append(append([]string{}, expectedCDKeys()...), collectNonCDRuntimeKeyNames(
		parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime,
	)...)
	for _, repoPath := range parsed.PublicSensitiveSurfaceGuard.KeyNameScopePaths {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, key := range collectRiidoAIServerKeyLiterals(body) {
			if !stringSliceContains(allowedPublicKeys, key) {
				t.Fatalf("%s contains unregistered public CD configuration key %q", repoPath, key)
			}
		}
	}
	assertBroadSummaryDocsLinkKeys(t, parsed, allowedPublicKeys)
}

func assertBroadSummaryDocsLinkKeys(t *testing.T, parsed runtimeCDOwnership, allowedPublicKeys []string) {
	t.Helper()
	nonCDNames := collectNonCDRuntimeKeyNames(parsed.PublicSensitiveSurfaceGuard.AllowedPublicNonCDRuntime)
	for _, repoPath := range parsed.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, key := range allowedPublicKeys {
			if stringSliceContains(nonCDNames, key) {
				continue
			}
			if strings.Contains(body, key) {
				t.Fatalf("%s must not repeat public CD configuration key %q", repoPath, key)
			}
		}
	}
}
