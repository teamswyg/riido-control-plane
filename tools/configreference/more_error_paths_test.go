package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigReferenceWriteTextRejectsBlockedParent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if err := writeText(filepath.Join(blocker, "doc.md"), "x"); err == nil {
		t.Fatal("expected writeText parent error")
	}
}

func TestConfigReferenceLoadManifestRejectsMalformedJSON(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":`), 0o644); err != nil {
		t.Fatalf("write malformed manifest: %v", err)
	}
	if _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestConfigReferenceVerifyAllRejectsInvalidGoSource(t *testing.T) {
	t.Parallel()
	repo := writeMiniRepo(t, "package main\nfunc broken(")
	m := testManifest("cmd/app", testEntry("RIIDO_ENV"))
	if _, err := verifyAll(repo, m); err == nil {
		t.Fatal("expected source parse error")
	}
}

func TestConfigReferenceRunRejectsMissingManifest(t *testing.T) {
	t.Parallel()
	repo := writeMiniRepo(t, `package main
func main() {}
`)
	err := run(options{Repo: repo, Manifest: "missing.json"})
	if err == nil || !strings.Contains(err.Error(), "open manifest") {
		t.Fatalf("expected open manifest error, got %v", err)
	}
}
