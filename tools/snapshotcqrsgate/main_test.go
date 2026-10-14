package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/snapshotcqrsgate/requirements"
)

func TestSnapshotCQRSGateBehaviorGolden(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot-gate.json")
	if err := run([]string{"-check-doc", "-evidence-out", path}); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.SchemaVersion != requirements.EvidenceSchema || got.ID != requirements.RequiredID || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.OperationsVerified != 3 || got.SignalsVerified != 7 {
		t.Fatalf("unexpected operation/signal evidence: %+v", got)
	}
	if got.DecisionRules != 2 || got.ForbiddenAttributes != 12 {
		t.Fatalf("unexpected rule/attribute evidence: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
}
