package main

import (
	"path/filepath"
	"testing"
)

func TestLoadFragmentsReturnsMissingFragmentError(t *testing.T) {
	m := minimalReadmeManifest()
	m.Fragments = []string{"missing.json"}
	if err := loadFragments(t.TempDir(), &m); err == nil {
		t.Fatal("expected missing fragment error")
	}
}

func TestLoadManifestMergesFragmentContent(t *testing.T) {
	root := t.TempDir()
	fragmentPath := filepath.Join(root, "fragment.json")
	writeReadmeTestFile(t, fragmentPath, `{
	  "schema_version":"riido-control-plane-repository-readme-fragment.v1",
	  "id":"fragment",
	  "summary":["fragment summary"],
	  "required_markers":["fragment summary"]
	}`)
	m := minimalReadmeManifest()
	m.Summary = nil
	m.RequiredMarkers = nil
	m.Fragments = []string{"fragment.json"}
	manifestPath := filepath.Join(root, "manifest.json")
	writeReadmeManifest(t, manifestPath, m)
	got, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Summary) != 1 || got.Summary[0] != "fragment summary" {
		t.Fatalf("summary not merged: %+v", got.Summary)
	}
}
