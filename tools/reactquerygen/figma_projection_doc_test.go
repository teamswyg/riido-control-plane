package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readFigmaProjectionDoc(t *testing.T, docPath string) string {
	t.Helper()
	doc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read projection doc: %v", err)
	}
	return string(doc)
}

func verifyFigmaProjectionDocumentBoundaries(t *testing.T, docText string) {
	t.Helper()
	if !strings.Contains(docText, "does not redefine the Figma top-level UI coverage") {
		t.Fatalf("projection doc must name the downstream-only boundary")
	}
}

func verifyFigmaProjectionStaleBoundaries(t *testing.T) {
	t.Helper()
	root := filepath.Join("..", "..")
	assertNoStaleFigmaNodeReference(t, filepath.Join(root, "docs"), "164-50215")
	assertNoStaleFigmaNodeReference(t, filepath.Join(root, "docs"), "164:50215")
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "docs"), "starter-agent")
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "docs"), "starter agent")
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "docs"), "future desktop or web clients")
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "docs"), "future desktop/web clients")
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "docs"), "future client bootstrap")
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "contracts", "ai-agent-client"), "future client bootstrap")
	staleRuntimeHost := "desktop-api." + "riido.ai"
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "docs"), staleRuntimeHost)
	assertNoStaleControlPlanePhrase(t, filepath.Join(root, "contracts", "ai-agent-client"), staleRuntimeHost)
}

func requireDocMentions(t *testing.T, docText, owner string, needles []string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(docText, needle) {
			t.Fatalf("projection doc must mention %s with %q", owner, needle)
		}
	}
}
