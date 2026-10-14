package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyRejectsContextMapPipelineCommandDrift(t *testing.T) {
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
	value["steps"].([]any)[8].(map[string]any)["command"] = contextDependencyNativeCommand
	raw, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("context map pipeline command substitution must fail closed")
	}
}
