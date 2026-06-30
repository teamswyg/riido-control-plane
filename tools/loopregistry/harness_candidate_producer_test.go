package main

import (
	"strings"
	"testing"
)

func TestHarnessCandidateProducerAcceptsOperationalReadiness(t *testing.T) {
	text := workflowTextForTest(t, ".github/workflows/operational-readiness.yml")
	if !harnessWorkflowProducesCandidates(text, "operational-readiness-closed-loop-candidates") {
		t.Fatal("operational readiness workflow must always produce candidate evidence")
	}
}

func TestHarnessCandidateProducerRequiresAlways(t *testing.T) {
	text := workflowTextForTest(t, ".github/workflows/operational-readiness.yml")
	broken := strings.Replace(text, "      - name: Generate operational readiness evidence\n        if: always()\n", "      - name: Generate operational readiness evidence\n", 1)
	if broken == text {
		t.Fatal("fixture did not contain always candidate producer")
	}
	if harnessWorkflowProducesCandidates(broken, "operational-readiness-closed-loop-candidates") {
		t.Fatal("expected candidate producer without if: always() to fail")
	}
}
