package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/openquestions/pathutil"
)

func TestOpenQuestionsBehaviorGolden(t *testing.T) {
	repo := filepath.Join("..", "..")
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := run(options{
		Repo:        repo,
		Manifest:    defaultManifest,
		EvidenceOut: out,
		CheckDoc:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	assertOpenQuestionsEvidence(t, got)
}

func assertOpenQuestionsEvidence(t *testing.T, got evidence) {
	t.Helper()
	if got.SchemaVersion != "riido-control-plane-open-questions-evidence.v1" {
		t.Fatalf("schema = %s", got.SchemaVersion)
	}
	if got.ID != "control-plane-open-questions" || got.Status != "verified" {
		t.Fatalf("identity/status = %s/%s", got.ID, got.Status)
	}
	if got.GeneratedDoc != "docs/50-roadmap/open-questions.md" ||
		got.Workflow != ".github/workflows/open-questions.yml" {
		t.Fatalf("doc/workflow = %s/%s", got.GeneratedDoc, got.Workflow)
	}
	if got.Result.QuestionCount != 8 || got.Result.OpenCount != 7 || got.Result.ResolvedCount != 1 || len(got.OpenCommands) != 7 {
		t.Fatalf("counts = %+v commands=%d", got.Result, len(got.OpenCommands))
	}
	if got.Result.StatusCounts["open"] != 7 ||
		got.Result.StatusCounts["resolved-no-diff"] != 1 {
		t.Fatalf("status counts = %+v", got.Result.StatusCounts)
	}
}

func TestOpenQuestionsBindOwnerNextArtifactAndReader(t *testing.T) {
	m, err := loadManifest(pathutil.RepoPath(filepath.Join("..", ".."), defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	if m.GeneratedDoc == "" {
		t.Fatal("generated reader doc is required")
	}
	for _, item := range m.Questions {
		if item.Owner == "" {
			t.Fatalf("question %s missing owner", item.ID)
		}
		if item.Status == "open" && item.NextArtifact == "" {
			t.Fatalf("open question %s missing next artifact", item.ID)
		}
		if item.Status == "open" && item.NextCommand == "" {
			t.Fatalf("open question %s missing next command", item.ID)
		}
	}
}
