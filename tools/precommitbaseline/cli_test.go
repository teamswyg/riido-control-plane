package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPreCommitBaselineMainRunWritesDocAndEvidence(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n")
	writeFile(t, filepath.Join(repo, ".pre-commit-config.yaml"), "id: go-fmt\n")
	writeFile(t, filepath.Join(repo, "scripts/check.sh"), "go test ./...\n")
	writeFile(t, filepath.Join(repo, ".github/workflows/precommit.yml"), workflowText())
	body, err := json.Marshal(manifest{
		SchemaVersion: manifestSchema, ID: "pre-commit", Title: "Pre Commit",
		GeneratedDoc: "docs/pre-commit.md", Workflow: ".github/workflows/precommit.yml",
		Evidence: "pre-commit-evidence", EvidenceTTL: 24,
		PreCommitConfig: ".pre-commit-config.yaml",
		Hooks: []checkBlock{{
			ID: "go-fmt", Summary: "Go fmt", Contains: []string{"id: go-fmt"},
		}},
		Scripts: []scriptSpec{{
			Path: "scripts/check.sh", Summary: "Go test", Contains: []string{"go test ./..."},
		}},
		Loop: completeLoop(),
	})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "baseline.json"), string(body))
	evidenceOut := filepath.Join(repo, "out/evidence.json")
	err = mainRun([]string{
		"-repo", repo, "-manifest", "baseline.json",
		"-write-doc", "-evidence-out", evidenceOut,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "docs/pre-commit.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(evidenceOut); err != nil {
		t.Fatal(err)
	}
}

func TestPreCommitBaselineMainRunRejectsBadFlag(t *testing.T) {
	if err := mainRun([]string{"-unknown"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}

func completeLoop() loopRecord {
	return loopRecord{"observe", "hypothesis", "execute", "evaluate", "retrospective"}
}

func workflowText() string {
	return "on:\n  schedule:\n    - cron: \"43 20 * * *\"\n" +
		"uses: actions/upload-artifact@v7\nname: pre-commit-evidence\n" +
		"if-no-files-found: error\n"
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
