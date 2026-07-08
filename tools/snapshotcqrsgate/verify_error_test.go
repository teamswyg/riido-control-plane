package main

import "testing"

func TestVerifyRejectsHeaderAndLoopDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"identity": func(m *manifest) { m.ID = "other" },
		"human":    func(m *manifest) { m.HumanDoc = "other.md" },
		"binding":  func(m *manifest) { m.GeneratedDoc = "other.md" },
		"scope":    func(m *manifest) { m.Decision.StoreWideCQRS = true },
		"loop":     func(m *manifest) { m.Loop.Evaluate = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := snapshotGateFixture()
			mutate(&m)
			if _, err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected verify error")
			}
		})
	}
}

func TestVerifyRejectsOperationAndSignalDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"operation-shape": func(m *manifest) { m.OperationEvidence[0].Route = "" },
		"operation-loss":  func(m *manifest) { m.OperationEvidence[1].StoreOperations = nil },
		"signal-loss":     func(m *manifest) { m.MeasurementSignals = []string{"X-Ray"} },
	} {
		t.Run(name, func(t *testing.T) {
			m := snapshotGateFixture()
			mutate(&m)
			if _, err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected operation or signal error")
			}
		})
	}
}

func TestVerifyRejectsDecisionAndTraceDrift(t *testing.T) {
	for name, mutate := range map[string]func(*manifest){
		"few-rules":   func(m *manifest) { m.DecisionRules = m.DecisionRules[:1] },
		"threshold":   func(m *manifest) { m.DecisionRules[0].ThresholdDropPercent = 1 },
		"action":      func(m *manifest) { m.DecisionRules[0].Action = "other" },
		"cadence":     func(m *manifest) { m.CadenceRules[0].Seconds = 20 },
		"split":       func(m *manifest) { m.CandidateSplit.QueryModels = nil },
		"trace-attrs": func(m *manifest) { m.ForbiddenTraceAttributes = []string{"task_id"} },
	} {
		t.Run(name, func(t *testing.T) {
			m := snapshotGateFixture()
			mutate(&m)
			if _, err := verify(t.TempDir(), m, false); err == nil {
				t.Fatalf("expected decision or trace error")
			}
		})
	}
}
