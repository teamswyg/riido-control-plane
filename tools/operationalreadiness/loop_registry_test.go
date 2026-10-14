package main

import "testing"

func TestOperationalReadinessBindsLoopRegistryExpiry(t *testing.T) {
	m := loadManifestForTest(t)
	registry := readinessLoopRegistry{Loops: []readinessRegistryLoop{readinessLoopFixture(m)}}
	if err := verifyRegisteredReadinessLoop(m, registry); err != nil {
		t.Fatal(err)
	}
	registry.Loops[0].ExpiresAfterHours = readinessEvidenceTTLHours + 1
	if err := verifyRegisteredReadinessLoop(m, registry); err == nil {
		t.Fatal("expected loop registry expiry drift to fail")
	}
}

func TestOperationalReadinessBindsLoopRegistryArtifacts(t *testing.T) {
	m := loadManifestForTest(t)
	loop := readinessLoopFixture(m)
	loop.Evidence = loop.Evidence[:1]
	registry := readinessLoopRegistry{Loops: []readinessRegistryLoop{loop}}
	if err := verifyRegisteredReadinessLoop(m, registry); err == nil {
		t.Fatal("expected missing loop registry artifact to fail")
	}
}

func TestOperationalReadinessRequiresHarnessPromotionLoop(t *testing.T) {
	m := loadManifestForTest(t)
	loop := readinessLoopFixture(m)
	loop.Kind = "closed_loop"
	registry := readinessLoopRegistry{Loops: []readinessRegistryLoop{loop}}
	if err := verifyRegisteredReadinessLoop(m, registry); err == nil {
		t.Fatal("expected readiness loop kind drift to fail")
	}
}

func readinessLoopFixture(m manifest) readinessRegistryLoop {
	return readinessRegistryLoop{
		ID:                readinessHarnessLoop,
		Kind:              "harness",
		RefreshWorkflow:   m.Workflow,
		ExpiresAfterHours: readinessEvidenceTTLHours,
		PromotesTo:        []string{"closed_loop_candidate"},
		Evidence: []readinessRegistryEvidence{
			{Path: m.EvidenceArtifact},
			{Path: readinessCandidateArtifact},
			{Path: m.GeneratedDoc},
		},
	}
}
