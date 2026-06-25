package main

func pressureAdoptionPlan(scenario string) []adoptionStep {
	return []adoptionStep{
		{"claim_binding", "update docs/30-architecture/loop-registry.riido.json for " + scenario},
		{"verifier", "add or extend a focused verifier for " + scenario},
		{"ci_gate", "bind the verifier to a lightweight CI or pre-commit gate"},
		{"redacted_evidence", "publish before/after pressure evidence for " + scenario},
		{"decision_record", "record the observation, hypothesis, decision, and next loop"},
		{"evidence_graph_edge", "update docs/30-architecture/evidence-graph.riido.json"},
	}
}
