package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRejectsPipelineOrderOrCommandDrift(t *testing.T) {
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
	value["steps"].([]any)[2].(map[string]any)["command"] = "go test ./..."
	raw, err = json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("pipeline command substitution must fail closed")
	}
}

func TestVerifyRejectsRunnerPinOrCredentialBoundaryDrift(t *testing.T) {
	repo := copyFixtureRepo(t)
	path := filepath.Join(repo, "tools", "riido-ci-local")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw = []byte(strings.Replace(string(raw), "4795cdebfcf4bccaa71ec2344368ad9adf6b1974", "1dd3ace14bf5090c478abb4e724aed91ae468b6c", 1))
	if err := os.WriteFile(path, raw, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := verify(repo, repositoryContract); err == nil {
		t.Fatal("runner revision drift must fail closed")
	}
	document := readRepositoryManifest(t)
	document.Runner.GitHubTokenRequired = true
	if _, err := verify(writeManifest(t, document), repositoryContract); err == nil {
		t.Fatal("GitHub token admission must fail closed")
	}
}
