package main

func pressureAdoptionPlan(scenario string) []adoptionStep {
	return []adoptionStep{
		{"claim_binding", "go run ./tools/loopregistry -check-doc -github-annotations"},
		{"verifier", "go test ./tools/controlplanepressure ./tools/controlplaneperf ./tools/controlplaneaudit -count=1"},
		{"ci_gate", "go run ./tools/controlplaneperf -check-doc -evidence-out out/control-plane-performance-evidence.json"},
		{"redacted_evidence", "go run ./tools/controlplanepressure -duration 500ms -concurrency 1,8,32 -threads 24 -lines 40 -evidence-out out/control-plane-local-pressure.json"},
		{"decision_record", "go run ./tools/loopclosureaudit -candidate-out out/loop-closure-audit-closed-loop-candidates.json"},
		{"evidence_graph_edge", "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph-evidence.json"},
	}
}
