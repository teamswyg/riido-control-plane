package main

const (
	pressureHarnessLoop     = "control_plane_performance_harness"
	pressurePromotionTarget = "closed_loop_candidate"
)

func candidateForScenario(sc scenario) candidateEntry {
	return candidateEntry{
		ID:                    "control-plane-pressure:" + sc.name,
		HarnessLoop:           pressureHarnessLoop,
		PromotionTarget:       pressurePromotionTarget,
		Scenario:              sc.name,
		Risk:                  sc.risk,
		Next:                  sc.next,
		RequiredNextArtifacts: pressureCandidateArtifacts(),
		AdoptionPlan:          pressureAdoptionPlan(sc.name),
	}
}

func pressureCandidateArtifacts() []string {
	return []string{
		"claim_binding",
		"verifier",
		"ci_gate",
		"redacted_evidence",
		"decision_record",
		"evidence_graph_edge",
	}
}
