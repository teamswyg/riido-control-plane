package main

import (
	"fmt"
	"strings"
)

func verifyAll(root string, m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != "control-plane-high-traffic-audit" {
		return fmt.Errorf("unexpected high-traffic audit manifest identity")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceTool != "tools/controlplaneaudit" {
		return fmt.Errorf("audit manifest must bind doc, workflow, and tool")
	}
	if m.EvidenceArtifact == "" || len(m.Surfaces) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("audit manifest must bind artifact, surfaces, and assertions")
	}
	if len(m.RequiredCategories) == 0 {
		return fmt.Errorf("audit manifest must declare required categories")
	}
	if err := verifyCommands(m); err != nil {
		return err
	}
	if err := verifySurfaces(root, m.Surfaces); err != nil {
		return err
	}
	if err := verifyRequiredCategories(m.RequiredCategories, m.Surfaces); err != nil {
		return err
	}
	if err := verifyWorkflow(root, m); err != nil {
		return err
	}
	return verifyLoop(m.Loop)
}

func verifyCommands(m manifest) error {
	required := map[string]string{
		m.BenchmarkCommand:      "-benchmem",
		m.LocalPressureCommand:  "go run ./tools/controlplanepressure",
		m.ManualPressureCommand: "-concurrency 1,8,32,128",
		m.LocalPprofCommand:     "-pprof-dir",
		m.RaceCommand:           "go test -race",
	}
	for command, needle := range required {
		if !strings.Contains(command, needle) {
			return fmt.Errorf("audit command missing %q: %s", needle, command)
		}
	}
	for _, command := range []string{
		m.LocalPressureCommand,
		m.ManualPressureCommand,
		m.LocalPprofCommand,
	} {
		if !strings.Contains(command, "-candidate-out") {
			return fmt.Errorf("audit pressure command missing candidate output: %s", command)
		}
	}
	return verifyPprofCommands(m.PprofCommands)
}
