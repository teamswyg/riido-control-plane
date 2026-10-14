package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-control-plane/tools/snapshotcqrsgate/requirements"
)

func snapshotGateFixture() manifest {
	return manifest{
		SchemaVersion: requirements.ManifestSchema, ID: requirements.RequiredID,
		RiidoTask: requirements.RequiredTask, GeneratedDoc: requirements.RequiredHumanDoc,
		Workflow: requirements.Workflow, EvidenceArtifact: requirements.EvidenceArtifact,
		HumanDoc: requirements.RequiredHumanDoc,
		Decision: decision{Scope: "ai_agent_client_snapshot_only", Reason: "scoped"},
		OperationEvidence: []operationEvidence{
			{"runtime", "sync", []string{"ai_agent_client_snapshot_load", "ai_agent_client_snapshot_save"}},
			{"poll", "poll", []string{"store_poll_assignment"}},
		},
		MeasurementSignals: []string{
			"ai_agent_client_snapshot_load_calls_total", "ai_agent_client_snapshot_save_calls_total",
			"DynamoDB ConsumedReadCapacityUnits", "DynamoDB ConsumedWriteCapacityUnits", "X-Ray samples",
		},
		CadenceRules: []cadenceRule{{"reload", 15, 20}, {"save", 15, 20}},
		DecisionRules: []decisionRule{
			{"keep", requirements.MinDecisionThreshold, "when", "keep_monolithic_snapshot"},
			{"split", requirements.MinDecisionThreshold, "when", "split_ai_agent_client_snapshot_only"},
		},
		CandidateSplit:           candidateSplit{[]string{"command"}, []string{"query"}},
		ForbiddenTraceAttributes: []string{"task_id", "agent_id", "credentials", "payload_document"},
		Loop:                     evidenceLoop{"observe", "hypothesis", "execute", "evaluate", "retro"},
	}
}

func writeSnapshotGateRepo(t *testing.T, m manifest) string {
	t.Helper()
	repo := t.TempDir()
	writeSnapshotFile(t, filepath.Join(repo, "go.mod"), "module example.com/snapshot\n")
	writeSnapshotJSON(t, filepath.Join(repo, "manifest.json"), m)
	return repo
}

func writeSnapshotJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeSnapshotFile(t, path, string(body))
}

func writeSnapshotFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
