package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRejectDuplicatesAndInlineSecrets(t *testing.T) {
	c := fixture()
	c.Environment = append(c.Environment, c.Environment[0])
	if validateFixture(c) == nil {
		t.Fatal("duplicate")
	}
	c = fixture()
	c.Secrets = append(c.Secrets, c.Secrets[0])
	if validateFixture(c) == nil {
		t.Fatal("duplicate secret")
	}
	c = fixture()
	c.Environment = append(c.Environment, entry{Name: keyNames[0], Value: "private"})
	if validateFixture(c) == nil {
		t.Fatal("inline/secret overlap")
	}
}

func TestWorkflowRunsBeforeMutationAndRechecksCandidate(t *testing.T) {
	b, e := os.ReadFile("../../.github/workflows/deploy-ai-agent-testnet.yml")
	if e != nil {
		t.Fatal(e)
	}
	s := string(b)
	gate := strings.Index(s, "- name: Verify production native AI configuration")
	build := strings.Index(s, "- name: Build and push immutable image")
	if gate < 0 || gate > build {
		t.Fatal("preflight must precede image publication")
	}
	if strings.Count(s, "go run ./tools/ainativepreflight") != 2 {
		t.Fatal("must also validate exact copied definition before registration")
	}
}

func TestRejectLegacyContextAndOpaqueEnvironment(t *testing.T) {
	for _, key := range []string{"RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID", "RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID", "RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY"} {
		c := fixture()
		c.Environment = append(c.Environment, entry{Name: key, Value: "old-value"})
		if validateFixture(c) == nil {
			t.Fatal("legacy env passed")
		}
		c = fixture()
		c.Secrets = append(c.Secrets, entry{Name: key, ValueFrom: "/fixture/production/old-value"})
		if validateFixture(c) == nil {
			t.Fatal("legacy secret passed")
		}
	}
	c := fixture()
	c.EnvironmentFiles = []json.RawMessage{json.RawMessage(`{"type":"s3","value":"private"}`)}
	if validateFixture(c) == nil {
		t.Fatal("opaque environment passed")
	}
}
