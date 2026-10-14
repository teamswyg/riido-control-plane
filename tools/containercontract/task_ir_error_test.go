package main

import "testing"

func TestVerifyTaskDefinitionIRRejectsInvalidTaskShape(t *testing.T) {
	for name, body := range map[string]string{
		"bad-json": "{",
		"trailing": fixtureTaskIR() + "{}",
		"os": `{"runtime_platform":{"operatingSystemFamily":"WINDOWS"},
			"container":{"portMappings":[{"containerPort":8080}],
			"environment":[{"name":"RIIDO_AI_SERVER_ADDR","value":":8080"}]}}`,
		"port": `{"runtime_platform":{"operatingSystemFamily":"LINUX"},
			"container":{"portMappings":[{"containerPort":9090}],
			"environment":[{"name":"RIIDO_AI_SERVER_ADDR","value":":8080"}]}}`,
		"env": `{"runtime_platform":{"operatingSystemFamily":"LINUX"},
			"container":{"portMappings":[{"containerPort":8080}],
			"environment":[{"name":"RIIDO_AI_SERVER_ADDR","value":":9090"}]}}`,
	} {
		t.Run(name, func(t *testing.T) {
			contract, _ := writeContainerContractFixture(t, "65532:65532", true)
			writeFixtureFile(t, contract.FargateTaskDefinitionIR, body)
			if err := verifyTaskDefinitionIR(contract); err == nil {
				t.Fatalf("expected task definition IR error")
			}
		})
	}
}

func TestLoadTaskDefinitionIRRejectsMissingFile(t *testing.T) {
	if _, err := loadTaskDefinitionIR("/tmp/riido-missing-task-ir.json"); err == nil {
		t.Fatalf("expected missing task IR error")
	}
}
