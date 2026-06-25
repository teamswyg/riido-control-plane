package main

import (
	"strings"
	"testing"
)

func TestCandidateDecisionRequiresSemanticWorkflowPathTriggers(t *testing.T) {
	m := manifest{
		Workflow:             ".github/workflows/decision.yml",
		GeneratedDoc:         "docs/decision.md",
		IntakeManifest:       "docs/intake.riido.json",
		LoopRegistryManifest: "docs/loop-registry.riido.json",
	}
	intake := intakeManifest{Sources: []intakeSource{{
		ProducerManifest:     "docs/producer.riido.json",
		LoopRegistryManifest: "docs/loop-registry.riido.json",
	}}}
	text := strings.Join([]string{
		m.Workflow,
		m.GeneratedDoc,
		m.IntakeManifest,
		m.LoopRegistryManifest,
	}, "\n")
	err := verifyWorkflowPathTriggers(text, m, intake)
	if err == nil || !strings.Contains(err.Error(), intake.Sources[0].ProducerManifest) {
		t.Fatalf("expected missing producer manifest path trigger, got %v", err)
	}
}
