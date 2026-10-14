package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigReferenceRunWritesDocAndEvidence(t *testing.T) {
	t.Parallel()
	source := `package main
import "os"
func main() { _ = os.Getenv("RIIDO_ENV") }
`
	repo := writeMiniRepo(t, source)
	manifestPath := filepath.Join(repo, "manifest.json")
	m := testManifest("cmd/app", testEntry("RIIDO_ENV"))
	writeManifestFile(t, manifestPath, m)
	evidencePath := filepath.Join(repo, "out", "evidence.json")
	err := run(options{Repo: repo, Manifest: manifestPath, EvidenceOut: evidencePath, WriteDoc: true, CheckDoc: true})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, m.GeneratedDoc)); err != nil {
		t.Fatalf("generated doc missing: %v", err)
	}
	if body, err := os.ReadFile(evidencePath); err != nil || !strings.Contains(string(body), `"status": "verified"`) {
		t.Fatalf("weak evidence output: %s/%v", body, err)
	}
}

func TestConfigReferenceRunRejectsStaleDoc(t *testing.T) {
	t.Parallel()
	repo := writeMiniRepo(t, `package main
const env = "RIIDO_ENV"
func main() { _ = env }
`)
	manifestPath := filepath.Join(repo, "manifest.json")
	m := testManifest("cmd/app", testEntry("RIIDO_ENV"))
	writeManifestFile(t, manifestPath, m)
	docPath := filepath.Join(repo, m.GeneratedDoc)
	if err := os.MkdirAll(filepath.Dir(docPath), 0o755); err != nil {
		t.Fatalf("mkdir doc: %v", err)
	}
	if err := os.WriteFile(docPath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale doc: %v", err)
	}
	err := run(options{Repo: repo, Manifest: manifestPath, CheckDoc: true})
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale doc error, got %v", err)
	}
}
