package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func fixture() container {
	c := container{Name: "ai"}
	for k, v := range map[string]string{"RIIDO_AI_SERVER_TASK_CONTEXT_LOCATOR_VERSION": "2", "RIIDO_AI_SERVER_ASSIGNMENT_READ_SCOPE": "true"} {
		c.Environment = append(c.Environment, entry{Name: k, Value: v})
	}
	for k, path := range urlPaths {
		c.Environment = append(c.Environment, entry{Name: k, Value: "https://monolith.example.com" + path})
	}
	for _, k := range keyNames {
		c.Secrets = append(c.Secrets, entry{Name: k, ValueFrom: "/fixture/production/" + k})
	}
	return c
}
func validateFixture(c container) error {
	b, _ := json.Marshal(task{Containers: []container{c}})
	return validate(b, "ai")
}
func TestValidWrappedAndBare(t *testing.T) {
	c := fixture()
	if e := validateFixture(c); e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(map[string]any{"taskDefinition": task{Containers: []container{c}}})
	if e := validate(b, "ai"); e != nil {
		t.Fatal(e)
	}
}
func TestRejectMissingAndUnsafeSettings(t *testing.T) {
	c := fixture()
	for i := range c.Environment {
		bad := fixture()
		bad.Environment = append(append([]entry{}, c.Environment[:i]...), c.Environment[i+1:]...)
		if validateFixture(bad) == nil {
			t.Fatal("accepted missing", c.Environment[i].Name)
		}
	}
	for i := range c.Secrets {
		bad := c
		bad.Secrets = append(append([]entry{}, c.Secrets[:i]...), c.Secrets[i+1:]...)
		if validateFixture(bad) == nil {
			t.Fatal("accepted missing key")
		}
	}
	for _, badURL := range []string{"http://monolith.example.com", "https://staging.example.com", "https://user:private@monolith.example.com", "https://monolith.example.com?token=private", "https://monolith.example.com/wrong", "https://monolith.example.com/#fragment"} {
		bad := fixture()
		for i := range bad.Environment {
			if bad.Environment[i].Name == "RIIDO_AI_SERVER_TASK_CONTEXT_BASE_URL" {
				bad.Environment[i].Value = badURL
			}
		}
		err := validateFixture(bad)
		if err == nil {
			t.Fatal("accepted unsafe endpoint")
		}
		if strings.Contains(err.Error(), badURL) || strings.Contains(err.Error(), "private") {
			t.Fatal("leaked input")
		}
	}
}
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
