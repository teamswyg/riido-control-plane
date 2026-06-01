package deploypolicy

import (
	"os"
	"strings"
	"testing"
)

func TestDeployAIAgentTestnetPublicRedactionPolicy(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/deploy-ai-agent-testnet.yml")
	readme := mustRead(t, "../../README.md")
	boundary := mustRead(t, "../../docs/30-architecture/runtime-deployment-boundary.md")
	domain := mustRead(t, "../../docs/20-domain/saas-control-plane.md")
	migration := mustRead(t, "../../docs/migration/control-plane.md")

	requireContains(t, workflow, "echo \"::add-mask::$aws_account_id\"")
	requireContains(t, workflow, "echo \"::add-mask::$registry\"")
	requireContains(t, workflow, "echo \"::add-mask::$image_uri\"")
	requireContains(t, workflow, "echo \"::add-mask::$current_task_definition\"")
	requireContains(t, workflow, "echo \"::add-mask::$next_task_definition\"")
	requireContains(t, workflow, "if [ \"$image_tag\" = \"latest\" ]")
	requireContains(t, workflow, "workflow_dispatch")
	requireContains(t, workflow, "tags:")
	requireContains(t, workflow, "- \"v*\"")

	for _, forbidden := range []string{
		"actions/upload-artifact",
		"latest\" >>",
		"latest' >>",
		"task-definition.next.json\" >>",
		"task-definition.current.json\" >>",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("deploy workflow must not contain %q", forbidden)
		}
	}

	requireContains(t, readme, "live URL, AWS account id, ARN, image digest")
	requireContains(t, boundary, "task-definition ARNs, image digests, live workflow run URLs")
	requireContains(t, boundary, "must not upload")
	requireContains(t, boundary, "deployment artifacts from the live run")
	requireContains(t, domain, "live URLs, task-definition ARNs")
	requireContains(t, migration, "RIID-4812 tightens that public boundary")

	for path, body := range map[string]string{
		"README.md":                      readme,
		"runtime-deployment-boundary.md": boundary,
		"saas-control-plane.md":          domain,
		"control-plane.md":               migration,
	} {
		if strings.Contains(body, "http://ai-api.riido.io") {
			t.Fatalf("%s must not pin the live testnet URL", path)
		}
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("missing %q", want)
	}
}
