package main

import (
	"path/filepath"
	"testing"
)

func TestOpenQuestionsManifest(t *testing.T) {
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
}

func TestOpenQuestionsBindOwnerNextArtifactAndReader(t *testing.T) {
	repo := filepath.Join("..", "..")
	m, err := loadManifest(repoPath(repo, defaultManifest))
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

func TestOpenQuestionsEvidenceExposesNextCommands(t *testing.T) {
	m, err := loadManifest(repoPath(filepath.Join("..", ".."), defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyAll(m)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result)
	if len(got.OpenCommands) != result.OpenCount {
		t.Fatalf("open command count = %d, want %d", len(got.OpenCommands), result.OpenCount)
	}
}
