package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs repo: %w", err)
	}
	for {
		if exists(filepath.Join(dir, "go.mod")) || exists(filepath.Join(dir, ".git")) {
			return dir, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("repo root not found from %s", start)
		}
		dir = next
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkGeneratedDoc(root, expected string) error {
	actual, err := readText(repoPath(root, defaultDoc))
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s is stale; run go run ./tools/figmaprojection -write-doc", defaultDoc)
	}
	return nil
}

func checkProjectionGate(root string) error {
	cmd := exec.Command("go", "test", "./tools/reactquerygen", "-run",
		"TestFigmaAIAgentControlPlaneProjectionManifest", "-count=1")
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("figma projection gate failed: %w\n%s", err, out.String())
	}
	return nil
}
