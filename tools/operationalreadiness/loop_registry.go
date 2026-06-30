package main

import "fmt"

type readinessLoopRegistry struct {
	Loops []readinessRegistryLoop `json:"loops"`
}

type readinessRegistryLoop struct {
	ID                string                      `json:"id"`
	RefreshWorkflow   string                      `json:"refresh_workflow"`
	ExpiresAfterHours int                         `json:"expires_after_hours"`
	Evidence          []readinessRegistryEvidence `json:"evidence"`
}

type readinessRegistryEvidence struct {
	Path string `json:"path"`
}

func verifyLoopRegistryBinding(root string, m manifest) error {
	if m.LoopRegistry == "" {
		return fmt.Errorf("readiness manifest must bind loop registry manifest")
	}
	var registry readinessLoopRegistry
	if err := readJSON(repoPath(root, m.LoopRegistry), &registry); err != nil {
		return err
	}
	return verifyRegisteredReadinessLoop(m, registry)
}

func verifyRegisteredReadinessLoop(m manifest, registry readinessLoopRegistry) error {
	source := readinessCandidateSource(m)
	for _, loop := range registry.Loops {
		if loop.ID == source.HarnessLoop {
			return verifyReadinessLoopSurface(m, loop)
		}
	}
	return fmt.Errorf("readiness loop %s missing in loop registry", source.HarnessLoop)
}

func verifyReadinessLoopSurface(m manifest, loop readinessRegistryLoop) error {
	if loop.RefreshWorkflow != m.Workflow {
		return fmt.Errorf("readiness loop refresh workflow drift")
	}
	if loop.ExpiresAfterHours != readinessEvidenceTTLHours {
		return fmt.Errorf("readiness loop expiry drift")
	}
	for _, path := range []string{m.EvidenceArtifact, readinessCandidateArtifact, m.GeneratedDoc} {
		if !readinessLoopHasEvidence(loop, path) {
			return fmt.Errorf("readiness loop evidence missing %s", path)
		}
	}
	return nil
}

func readinessLoopHasEvidence(loop readinessRegistryLoop, path string) bool {
	for _, evidence := range loop.Evidence {
		if evidence.Path == path {
			return true
		}
	}
	return false
}
