package main

import (
	"strings"
	"testing"
)

func TestVerifyContractPassesWithoutPrivateInfraIR(t *testing.T) {
	contract, _ := writeContainerContractFixture(t, "65532:65532", false)

	record, err := verifyContract(contract)
	if err != nil {
		t.Fatalf("verifyContract() error = %v", err)
	}
	if record.SchemaVersion != checkSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", record.SchemaVersion, checkSchemaVersion)
	}
	if record.ID == "" || record.Status != "verified" || record.Loop.Observation == "" {
		t.Fatalf("missing evidence metadata: %+v", record)
	}
	if record.Service != "riido_ai_server" {
		t.Fatalf("service = %q", record.Service)
	}
	if record.FinalBaseImage != "scratch" {
		t.Fatalf("final_base_image = %q", record.FinalBaseImage)
	}
	if record.FinalUser != "65532:65532" {
		t.Fatalf("final_user = %q", record.FinalUser)
	}
	if record.ChecksTotal != 17 {
		t.Fatalf("checks_total = %d, want 17", record.ChecksTotal)
	}
}

func TestVerifyContractPassesAndChecksOptionalTaskIR(t *testing.T) {
	contract, _ := writeContainerContractFixture(t, "65532:65532", true)

	record, err := verifyContract(contract)
	if err != nil {
		t.Fatalf("verifyContract() error = %v", err)
	}
	if record.FargateTaskDefinitionIR == "" {
		t.Fatal("fargate_task_definition_ir is empty")
	}
	if record.ChecksTotal != 20 {
		t.Fatalf("checks_total = %d, want 20", record.ChecksTotal)
	}
}

func TestVerifyContractRejectsRootUser(t *testing.T) {
	contract, _ := writeContainerContractFixture(t, "0:0", false)

	_, err := verifyContract(contract)
	if err == nil || !strings.Contains(err.Error(), "non-root") {
		t.Fatalf("verifyContract() error = %v, want non-root failure", err)
	}
}
