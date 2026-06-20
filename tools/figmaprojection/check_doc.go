package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

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
