package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func generateOperationalReadinessCandidate(t *testing.T, root, out string) error {
	t.Helper()
	cmd := exec.Command("go", "run", "./tools/operationalreadiness",
		"-candidate-out", out)
	cmd.Dir = root
	cmd.Env = append(os.Environ(),
		"RIIDO_OPERATIONAL_READINESS_NOW=2026-06-29T12:00:00Z")
	if body, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, body)
	}
	return nil
}
