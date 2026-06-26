package main

import "fmt"

func verifyPressureSources(m manifest) error {
	if len(m.Sources) == 0 {
		return fmt.Errorf("performance manifest must declare pressure candidate sources")
	}
	for _, source := range m.Sources {
		if source.ID == "control-plane-pressure" {
			return verifyPressureSource(m, source)
		}
	}
	return fmt.Errorf("performance manifest missing control-plane-pressure source")
}

func verifyPressureSource(m manifest, source pressureSource) error {
	if source.CandidateArtifact != m.PressureCandidateArtifact ||
		source.SummaryArtifact != m.LocalPressureArtifact ||
		source.SourceWorkflow != m.Workflow {
		return fmt.Errorf("control-plane-pressure source does not match performance artifacts")
	}
	if len(source.RequiredNextArtifacts) == 0 {
		return fmt.Errorf("control-plane-pressure source must declare required next artifacts")
	}
	return nil
}
