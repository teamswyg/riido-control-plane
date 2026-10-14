package deploypolicy

import "testing"

func TestDeployAIAgentTestnetPublicRedactionPolicy(t *testing.T) {
	fixture := loadRedactionFixture(t)
	assertDeployWorkflowRedaction(t, fixture.Workflow)
	assertSmokeWorkflowRedaction(t, fixture.SmokeWorkflow)
	assertPublicRedactionDocs(t, fixture)
	assertNoPinnedLiveHost(t, fixture)
}

func assertDeployWorkflowRedaction(t *testing.T, workflow string) {
	t.Helper()
	for _, want := range deployWorkflowRequiredPhrases() {
		requireContains(t, workflow, want)
	}
	for _, want := range deployRuntimeRequiredPhrases() {
		requireContains(t, workflow, want)
	}
	for _, masked := range deployMaskedKeys() {
		requireContains(t, workflow, masked)
	}
	requireRedactedSummaryArtifact(t, workflow, "deploy-ai-agent-testnet-redacted-summary")
	assertWorkflowForbiddenPhrases(t, "deploy", workflow)
}

func assertSmokeWorkflowRedaction(t *testing.T, workflow string) {
	t.Helper()
	for _, want := range smokeWorkflowRequiredPhrases() {
		requireContains(t, workflow, want)
	}
	requireRedactedSummaryArtifact(t, workflow, "ai-agent-client-testnet-smoke-redacted-summary")
	assertWorkflowForbiddenPhrases(t, "smoke", workflow)
}

func assertWorkflowForbiddenPhrases(t *testing.T, name, workflow string) {
	t.Helper()
	for _, forbidden := range workflowForbiddenPhrases() {
		if contains(workflow, forbidden) {
			t.Fatalf("%s workflow must not contain %q", name, forbidden)
		}
	}
}
