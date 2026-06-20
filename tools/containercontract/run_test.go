package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesCheckJSON(t *testing.T) {
	contract, contractPath := writeContainerContractFixture(t, "65532:65532", false)
	var out bytes.Buffer

	if err := run([]string{"-contract", contractPath, "-out", "-"}, &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	var record checkRecord
	if err := json.Unmarshal(out.Bytes(), &record); err != nil {
		t.Fatalf("decode check JSON: %v\n%s", err, out.String())
	}
	if record.SchemaVersion != checkSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", record.SchemaVersion, checkSchemaVersion)
	}
	if record.Service != contract.Service {
		t.Fatalf("service = %q, want %q", record.Service, contract.Service)
	}
	if record.ChecksTotal != 13 {
		t.Fatalf("checks_total = %d, want 13", record.ChecksTotal)
	}
}

func TestRunWritesEvidenceOutJSON(t *testing.T) {
	_, contractPath := writeContainerContractFixture(t, "65532:65532", false)
	outPath := filepath.Join(t.TempDir(), "container-evidence.json")

	if err := run([]string{"-contract", contractPath, "-evidence-out", outPath}, io.Discard); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var record checkRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode check JSON: %v\n%s", err, string(data))
	}
	if record.SchemaVersion != checkSchemaVersion || record.ChecksTotal != 13 {
		t.Fatalf("unexpected evidence: %+v", record)
	}
	if record.ID == "" || record.Status != "verified" || record.Loop.Observation == "" {
		t.Fatalf("missing evidence metadata: %+v", record)
	}
}
