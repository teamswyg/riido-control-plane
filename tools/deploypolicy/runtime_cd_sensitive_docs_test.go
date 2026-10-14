package deploypolicy

import (
	"path/filepath"
	"strings"
	"testing"
)

func assertBroadDocsDoNotListCDKeys(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	for _, repoPath := range p.PublicSensitiveSurfaceGuard.BroadSummaryDocsMustLink {
		body := mustRead(t, filepath.Join("..", "..", filepath.FromSlash(repoPath)))
		for _, key := range expectedCDKeys() {
			if strings.Contains(body, key) {
				t.Fatalf("%s must link to runtime CD ownership instead of listing CD key %q", repoPath, key)
			}
		}
	}
}
