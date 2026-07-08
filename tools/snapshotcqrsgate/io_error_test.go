package main

import (
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsMalformedInput(t *testing.T) {
	tmp := t.TempDir()
	if _, err := loadManifest(filepath.Join(tmp, "missing.json")); err == nil {
		t.Fatalf("expected missing manifest error")
	}
	for name, body := range map[string]string{
		"unknown":  `{"unknown":true}`,
		"bad-json": `{`,
		"trailing": `{}` + "\n{}",
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(tmp, name+".json")
			writeSnapshotFile(t, path, body)
			if _, err := loadManifest(path); err == nil {
				t.Fatalf("expected load manifest error")
			}
		})
	}
}

func TestWritersReportFailures(t *testing.T) {
	m := snapshotGateFixture()
	tmp := t.TempDir()
	if err := writeEvidence(filepath.Join(tmp, "ok.json"), newEvidence(m, result{})); err != nil {
		t.Fatal(err)
	}
	if err := writeDocFile(tmp, m); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(tmp, "blocker")
	writeSnapshotFile(t, blocker, "file")
	if err := writeEvidence(filepath.Join(blocker, "x.json"), newEvidence(m, result{})); err == nil {
		t.Fatalf("expected write evidence error")
	}
	m.HumanDoc = "blocker/doc.md"
	if err := writeDocFile(tmp, m); err == nil {
		t.Fatalf("expected write doc error")
	}
}
