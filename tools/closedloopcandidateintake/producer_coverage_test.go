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

func withoutIntakeSource(sources []intakeSource, id string) []intakeSource {
	out := make([]intakeSource, 0, len(sources))
	for _, source := range sources {
		if source.ID != id {
			out = append(out, source)
		}
	}
	return out
}
