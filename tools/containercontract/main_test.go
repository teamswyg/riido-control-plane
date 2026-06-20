package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
	if record.ChecksTotal != 13 {
		t.Fatalf("checks_total = %d, want 13", record.ChecksTotal)
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
	if record.ChecksTotal != 16 {
		t.Fatalf("checks_total = %d, want 16", record.ChecksTotal)
	}
}

func TestVerifyContractRejectsRootUser(t *testing.T) {
	contract, _ := writeContainerContractFixture(t, "0:0", false)

	_, err := verifyContract(contract)
	if err == nil {
		t.Fatal("verifyContract() error = nil, want non-root failure")
	}
	if !strings.Contains(err.Error(), "non-root") {
		t.Fatalf("verifyContract() error = %v, want non-root failure", err)
	}
}

func TestRunWritesCheckJSON(t *testing.T) {
	contract, contractPath := writeContainerContractFixture(t, "65532:65532", false)
	body, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contractPath, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

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
	contract, contractPath := writeContainerContractFixture(t, "65532:65532", false)
	body, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contractPath, append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "container-evidence.json")
	if err := run([]string{"-contract", contractPath, "-evidence-out", outPath}, io.Discard); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	var record checkRecord
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
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

func writeContainerContractFixture(t *testing.T, finalUser string, includeTaskIR bool) (imageContract, string) {
	t.Helper()
	dir := t.TempDir()
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	taskIRPath := filepath.Join(dir, "riido_ai_server_fargate.riido.json")
	contractPath := filepath.Join(dir, "riido_ai_server_container.riido.json")
	dockerfile := strings.Join([]string{
		"ARG GO_IMAGE=golang:1.26",
		"FROM ${GO_IMAGE} AS build",
		"WORKDIR /src",
		"COPY go.mod go.sum ./",
		"RUN go mod download",
		"COPY cmd ./cmd",
		"COPY internal ./internal",
		"ENV CGO_ENABLED=0",
		`RUN go build -trimpath -ldflags="-s -w" -o /out/riido_ai_server ./cmd/riido_ai_server`,
		"FROM scratch",
		"COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt",
		"COPY --from=build /out/riido_ai_server /riido_ai_server",
		"EXPOSE 8080",
		"ENV RIIDO_AI_SERVER_ADDR=:8080",
		"USER " + finalUser,
		`ENTRYPOINT ["/riido_ai_server"]`,
		"",
	}, "\n")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0o644); err != nil {
		t.Fatal(err)
	}
	taskIR := `{
  "schema_version": "riido-aws-fargate-task-definition.v1",
  "family": "riido-ai-server",
  "runtime_platform": {
    "cpuArchitecture": "X86_64",
    "operatingSystemFamily": "LINUX"
  },
  "container": {
    "name": "riido_ai_server",
    "portMappings": [
      {
        "containerPort": 8080,
        "hostPort": 8080,
        "protocol": "tcp"
      }
    ],
    "environment": [
      {
        "name": "RIIDO_AI_SERVER_ADDR",
        "value": ":8080"
      }
    ]
  }
}
`
	if err := os.WriteFile(taskIRPath, []byte(taskIR), 0o644); err != nil {
		t.Fatal(err)
	}
	contract := imageContract{
		SchemaVersion: contractSchemaVersion,
		ID:            "test-container-contract",
		Service:       "riido_ai_server",
		Dockerfile:    dockerfilePath,
		Assertions:    []string{"image must be non-root"},
		Loop: evidenceLoop{
			Observation:   "test",
			Hypothesis:    "test",
			Execute:       "test",
			Evaluate:      "test",
			Retrospective: "test",
		},
		Build: buildContract{
			BuildArg:   buildArgContract{Name: "GO_IMAGE", Default: "golang:1.26"},
			StageName:  "build",
			Workdir:    "/src",
			CGOEnabled: "0",
			GoBuild: goBuildContract{
				Package:  "./cmd/riido_ai_server",
				Output:   "/out/riido_ai_server",
				Trimpath: true,
				LDFlags:  []string{"-s", "-w"},
			},
		},
		Final: finalContract{
			BaseImage:  "scratch",
			CopyFrom:   "build",
			CopySource: "/out/riido_ai_server",
			Binary:     "/riido_ai_server",
			RequiredCopies: []requiredCopyContract{
				{
					From:        "build",
					Source:      "/etc/ssl/certs/ca-certificates.crt",
					Destination: "/etc/ssl/certs/ca-certificates.crt",
				},
			},
			ExposedPorts: []int{8080},
			Env: map[string]string{
				"RIIDO_AI_SERVER_ADDR": ":8080",
			},
			User:       finalUser,
			Entrypoint: []string{"/riido_ai_server"},
		},
	}
	if includeTaskIR {
		contract.FargateTaskDefinitionIR = taskIRPath
	}
	return contract, contractPath
}
