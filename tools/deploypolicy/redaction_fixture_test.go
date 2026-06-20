package deploypolicy

import "testing"

type redactionFixture struct {
	Workflow             string
	SmokeWorkflow        string
	Readme               string
	Boundary             string
	Domain               string
	ClientAPI            string
	ClientDelivery       string
	Migration            string
	Generator            string
	GeneratedClient      string
	GeneratedReactClient string
}

func loadRedactionFixture(t *testing.T) redactionFixture {
	t.Helper()
	return redactionFixture{
		Workflow:             mustRead(t, "../../.github/workflows/deploy-ai-agent-testnet.yml"),
		SmokeWorkflow:        mustRead(t, "../../.github/workflows/ai-agent-client-testnet-smoke.yml"),
		Readme:               mustRead(t, "../../README.md"),
		Boundary:             mustRead(t, "../../docs/30-architecture/runtime-deployment-boundary.md"),
		Domain:               mustRead(t, "../../docs/20-domain/saas-control-plane.md"),
		ClientAPI:            mustRead(t, "../../docs/20-domain/ai-agent-client-api.md"),
		ClientDelivery:       mustRead(t, "../../docs/30-architecture/api-client-delivery.md"),
		Migration:            mustRead(t, "../../docs/migration/control-plane.md"),
		Generator:            mustRead(t, "../../tools/reactquerygen/main.go"),
		GeneratedClient:      mustRead(t, "../../web/generated/aiAgentClient.ts"),
		GeneratedReactClient: mustRead(t, "../../web/generated/aiAgentClient.react.ts"),
	}
}
