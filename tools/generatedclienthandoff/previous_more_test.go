package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPreviousManifestIgnoresMissingAndEmptyPath(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"", filepath.Join(t.TempDir(), "missing.ts")} {
		got, err := readPreviousManifest(path)
		if err != nil {
			t.Fatalf("readPreviousManifest(%q): %v", path, err)
		}
		if got.Available {
			t.Fatalf("manifest %q available = true, want false", path)
		}
	}
}

func TestReadPreviousManifestWrapsReadFailure(t *testing.T) {
	t.Parallel()
	_, err := readPreviousManifest(t.TempDir())
	assertErrorContains(t, err, "read previous manifest")
}

func TestParsePreviousManifestUnescapesLifecycleFields(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "manifest.ts")
	body := `export const x = {
  sourceCommit: 'commit\'one',
  operations: [
    { generatedPath: 'b.path', operationId: 'opB', method: 'POST', path: '/b', deprecated: true, lifecycle: 'sun\'set', replacement: 'a.path', removalHorizon: '2027-Q1' },
    { generatedPath: 'a.path', operationId: 'opA', method: 'GET', path: '/a', deprecated: false },
  ],
} as const;`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	got, err := readPreviousManifest(path)
	if err != nil {
		t.Fatalf("readPreviousManifest: %v", err)
	}
	if got.SourceCommit != "commit'one" || got.Operations[0].GeneratedPath != "a.path" {
		t.Fatalf("manifest = %+v", got)
	}
	if got.Operations[1].Lifecycle != "sun'set" || !got.Operations[1].Deprecated {
		t.Fatalf("operation = %+v", got.Operations[1])
	}
}
