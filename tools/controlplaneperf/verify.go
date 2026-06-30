package main

import "fmt"

func verifyAll(root string, m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID != "control-plane-performance" {
		return fmt.Errorf("unexpected performance manifest identity")
	}
	if m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceTool != "tools/controlplaneperf" {
		return fmt.Errorf("performance manifest must bind doc, workflow, and tool")
	}
	if m.EvidenceArtifact == "" || m.BenchmarkArtifact == "" || m.LocalPressureArtifact == "" ||
		m.SummaryArtifact == "" || m.CandidateArtifact == "" || m.PressureCandidateArtifact == "" ||
		m.ArchitectureQueryArtifact == "" {
		return fmt.Errorf("performance manifest must bind evidence, benchmark, pressure, summary, and candidate artifacts")
	}
	if err := verifyPressureSources(m); err != nil {
		return err
	}
	if len(m.HotPaths) == 0 || len(m.Assertions) == 0 {
		return fmt.Errorf("performance manifest must declare hot paths and assertions")
	}
	if err := verifyCommands(m); err != nil {
		return err
	}
	if err := verifyBenchmarkHistory(root, m); err != nil {
		return err
	}
	if err := verifyHotPaths(root, m.HotPaths); err != nil {
		return err
	}
	if err := verifyArchitectureComponents(root, m); err != nil {
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
