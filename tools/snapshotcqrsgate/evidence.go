package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-control-plane/tools/snapshotcqrsgate/requirements"
)

type evidence struct {
	SchemaVersion       string       `json:"schema_version"`
	ID                  string       `json:"id"`
	Status              string       `json:"status"`
	OperationsVerified  int          `json:"operations_verified"`
	SignalsVerified     int          `json:"signals_verified"`
	DecisionRules       int          `json:"decision_rules"`
	ForbiddenAttributes int          `json:"forbidden_trace_attributes"`
	Loop                evidenceLoop `json:"loop"`
}

func newEvidence(m manifest, r result) evidence {
	return evidence{requirements.EvidenceSchema, m.ID, "verified", r.Operations, r.Signals, r.DecisionRules, r.ForbiddenAttributes, m.Loop}
}

func writeEvidence(path string, value evidence) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}
