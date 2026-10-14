package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeContainerContractFixture(t *testing.T, finalUser string, includeTaskIR bool) (imageContract, string) {
	t.Helper()
	dir := t.TempDir()
	dockerfilePath := filepath.Join(dir, "Dockerfile")
	taskIRPath := filepath.Join(dir, "riido_ai_server_fargate.riido.json")
	contractPath := filepath.Join(dir, "riido_ai_server_container.riido.json")
	writeFixtureFile(t, dockerfilePath, fixtureDockerfile(finalUser))
	writeFixtureFile(t, taskIRPath, fixtureTaskIR())
	contract := fixtureContract(dockerfilePath, finalUser)
	if includeTaskIR {
		contract.FargateTaskDefinitionIR = taskIRPath
	}
	writeFixtureJSON(t, contractPath, contract)
	return contract, contractPath
}

func fixtureContract(dockerfilePath, finalUser string) imageContract {
	return imageContract{
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
		Build: fixtureBuildContract(),
		Final: fixtureFinalContract(finalUser),
	}
}

func writeFixtureJSON(t *testing.T, path string, value any) {
	t.Helper()
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, path, string(append(body, '\n')))
}

func writeFixtureFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
}
