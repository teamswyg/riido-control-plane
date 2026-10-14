package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyRejectsEvidenceGraphPipelineCommandDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, "pipelines", "control-plane.local-self-check.riido.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	value["steps"].([]any)[44].(map[string]any)["command"] = evidenceGraphTestSource
	raw, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("evidence-graph pipeline command substitution must fail closed")
	}
}

func TestVerifyRejectsEvidenceGraphPipelineArtifactDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, "pipelines", "control-plane.local-self-check.riido.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	value["steps"].([]any)[45].(map[string]any)["run_when"] = "always"
	raw, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("evidence-graph artifact scheduling drift must fail closed")
	}
}
