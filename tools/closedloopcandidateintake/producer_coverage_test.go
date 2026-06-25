package main

import "testing"

func TestCandidateIntakeRequiresEveryProducerSource(t *testing.T) {
	root := repoRootForTest(t)
	m := loadIntakeManifestForTest(t)
	m.Sources = withoutIntakeSource(m.Sources, "control-plane-performance")
	if _, err := verifyAll(root, m); err == nil {
		t.Fatal("expected missing producer source coverage to fail")
	}
}

func TestCandidateIntakeSourceWorkflowMustMatchProducer(t *testing.T) {
	source := intakeSource{
		ID:                "control-plane-performance",
		CandidateArtifact: "control-plane-performance-closed-loop-candidates",
		SourceWorkflow:    ".github/workflows/wrong.yml",
		HarnessLoop:       "control_plane_performance_harness",
		PromotionTarget:   "closed_loop_candidate",
	}
	producer := producerSource{
		ID:                source.ID,
		CandidateArtifact: source.CandidateArtifact,
		SourceWorkflow:    ".github/workflows/control-plane-performance.yml",
		HarnessLoop:       source.HarnessLoop,
		PromotionTarget:   source.PromotionTarget,
	}
	if err := verifyProducerSource(source, producer); err == nil {
		t.Fatal("expected producer source workflow drift to fail")
	}
}

func withoutIntakeSource(sources []intakeSource, id string) []intakeSource {
	out := make([]intakeSource, 0, len(sources))
	for _, source := range sources {
		if source.ID != id {
			out = append(out, source)
		}
	}
	return out
}
