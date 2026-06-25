package main

import (
	"fmt"
	"strings"
)

func verifyAll(root string, m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != "control-plane-performance" {
		return fmt.Errorf("unexpected performance manifest identity")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceTool != "tools/controlplaneperf" {
		return fmt.Errorf("performance manifest must bind doc, workflow, and tool")
	}
	if m.EvidenceArtifact == "" || m.BenchmarkArtifact == "" || m.LocalPressureArtifact == "" ||
		m.SummaryArtifact == "" || m.CandidateArtifact == "" {
		return fmt.Errorf("performance manifest must bind evidence, benchmark, pressure, summary, and candidate artifacts")
	}
	if len(m.HotPaths) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("performance manifest must declare hot paths and assertions")
	}
	if err := verifyCommands(m); err != nil {
		return err
	}
	if err := verifyHotPaths(root, m.HotPaths); err != nil {
		return err
	}
	if err := verifyLocalPressureScenarios(root, m.LocalPressureScenarios); err != nil {
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
		m.RaceCommand:           "go test -race",
		m.PprofCommand:          "127.0.0.1:6060",
		m.LiveLoadCommand:       "go run ./tools/aiagentload",
	}
	for command, needle := range required {
		if !strings.Contains(command, needle) {
			return fmt.Errorf("performance command missing %q: %s", needle, command)
		}
	}
	return nil
}
