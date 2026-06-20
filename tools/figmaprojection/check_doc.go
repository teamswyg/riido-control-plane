package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func checkDoc(root string) error {
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
