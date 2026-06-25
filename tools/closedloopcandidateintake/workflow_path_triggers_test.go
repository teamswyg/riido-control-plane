package main

import (
	"strings"
	"testing"
)

func TestCandidateIntakeRequiresSemanticWorkflowPathTriggers(t *testing.T) {
	m := manifest{
		Workflow:     ".github/workflows/intake.yml",
		GeneratedDoc: "docs/intake.md",
		Sources: []intakeSource{{
			SourceWorkflow:        ".github/workflows/source.yml",
			ProducerManifest:      "docs/producer.riido.json",
			LoopRegistryManifest:  "docs/loop-registry.riido.json",
			EvidenceGraphManifest: "docs/evidence-graph.riido.json",
		}},
	}
	text := strings.Join([]string{
		m.Workflow,
		m.GeneratedDoc,
		m.Sources[0].SourceWorkflow,
		m.Sources[0].ProducerManifest,
		m.Sources[0].LoopRegistryManifest,
	}, "\n")
	err := verifyWorkflowPathTriggers(text, m)
	if err == nil || !strings.Contains(err.Error(), m.Sources[0].EvidenceGraphManifest) {
		t.Fatalf("expected missing evidence graph path trigger, got %v", err)
	}
}

func TestCandidateIntakeRequiresSourceWorkflowPathTrigger(t *testing.T) {
	m := manifest{Workflow: ".github/workflows/intake.yml"}
	m.Sources = []intakeSource{{SourceWorkflow: ".github/workflows/source.yml"}}
	err := verifyWorkflowPathTriggers(m.Workflow, m)
	if err == nil || !strings.Contains(err.Error(), m.Sources[0].SourceWorkflow) {
		t.Fatalf("expected missing source workflow path trigger, got %v", err)
	}
}
