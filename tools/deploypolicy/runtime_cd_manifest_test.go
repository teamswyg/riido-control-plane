package deploypolicy

import "testing"

func TestRuntimeCDOwnershipManifest(t *testing.T) {
	parsed := loadRuntimeCDOwnership(t)
	fixture := loadRuntimeCDDocFixture(t)
	assertRuntimeCDIdentity(t, parsed)
	assertRuntimeCDStrategies(t, parsed)
	assertRuntimeCDRedactionAndHandoff(t, parsed)
	assertRuntimeCDInfra(t, parsed)
	assertRuntimeCDPublicExport(t, parsed)
	assertRuntimeCDConfigKeys(t, parsed, fixture)
	assertRuntimeCDSensitiveSurface(t, parsed)
	assertRuntimeCDOperationalDetail(t, parsed)
	assertRuntimeCDDocs(t, parsed, fixture)
}

type runtimeCDDocFixture struct {
	Doc            string
	Boundary       string
	Integration    string
	Readme         string
	Domain         string
	Migration      string
	DeployWorkflow string
	SmokeWorkflow  string
}

func loadRuntimeCDDocFixture(t *testing.T) runtimeCDDocFixture {
	t.Helper()
	return runtimeCDDocFixture{
		Doc:            mustRead(t, "../../docs/30-architecture/runtime-cd-ownership.md"),
		Boundary:       mustRead(t, "../../docs/30-architecture/runtime-deployment-boundary.md"),
		Integration:    mustRead(t, "../../docs/30-architecture/integration-matrix.md"),
		Readme:         mustRead(t, "../../README.md"),
		Domain:         mustRead(t, "../../docs/20-domain/saas-control-plane.md"),
		Migration:      mustRead(t, "../../docs/migration/control-plane.md"),
		DeployWorkflow: mustRead(t, "../../.github/workflows/deploy-ai-agent-testnet.yml"),
		SmokeWorkflow:  mustRead(t, "../../.github/workflows/ai-agent-client-testnet-smoke.yml"),
	}
}
