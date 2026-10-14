package deploypolicy

import "testing"

func TestGeneratedClientDeliveryTokenBoundary(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/generated-client-delivery.yml")
	clientDelivery := mustRead(t, "../../docs/30-architecture/api-client-delivery.md")
	migration := mustRead(t, "../../docs/migration/control-plane.md")

	for _, want := range generatedClientWorkflowPhrases() {
		requireContains(t, workflow, want)
	}
	requireNotContains(t, workflow, "RIIDO_CLIENT_DELIVERY_TOKEN is required to open or update teamswyg/riido-client PRs.")
	requireNotContains(t, workflow, "react-query-")
	for _, want := range generatedClientDocPhrases() {
		requireContains(t, clientDelivery, want)
	}
	for _, want := range generatedClientMigrationPhrases() {
		requireContains(t, migration, want)
	}
}

func generatedClientWorkflowPhrases() []string {
	return []string{
		"actions/create-github-app-token@v1",
		"RIIDO_CLIENT_DELIVERY_APP_ID",
		"RIIDO_CLIENT_DELIVERY_PRIVATE_KEY",
		"RIIDO_CLIENT_DELIVERY_TOKEN",
		"Generated client delivery needs cross-repository write permission",
		"github.event.inputs.create_pr == 'true'",
		"target_branch must be the Riido work branchName",
		"grep -Eq '^[A-Z][A-Z0-9]*-[0-9]+-.+'",
	}
}
