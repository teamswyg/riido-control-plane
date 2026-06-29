package deploypolicy

import "testing"

func assertRuntimeCDIdentity(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	if p.SchemaVersion != "riido-control-plane-runtime-cd-ownership.v1" {
		t.Fatalf("unexpected schema version: %q", p.SchemaVersion)
	}
	if p.ID != "runtime-cd-ownership" || p.RiidoTask != "RIID-4825" || p.Runtime != "riido_ai_server" {
		t.Fatalf("manifest identity drifted: %#v", p)
	}
	if p.Loop.Observation == "" || p.Loop.Retrospective == "" {
		t.Fatalf("runtime CD ownership loop is missing: %#v", p.Loop)
	}
	for _, task := range requiredHardeningTasks() {
		requireSliceContains(t, p.Hardening, task)
	}
}

func requiredHardeningTasks() []string {
	return []string{
		"RIID-4833",
		"RIID-4835",
		"RIID-4836",
		"RIID-4837",
		"RIID-4839",
		"RIID-4842",
		"RIID-4844",
		"RIID-4845",
		"RIID-4853",
		"RIID-4855",
	}
}

func assertRuntimeCDCurrentStrategy(t *testing.T, current currentStrategy) {
	t.Helper()
	if current.CDOwner != "riido-control-plane" || current.TopologyOwner != "riido-infra" {
		t.Fatalf("current CD ownership drifted: %#v", current)
	}
	if current.Workflow != ".github/workflows/deploy-ai-agent-testnet.yml" {
		t.Fatalf("workflow drifted: %q", current.Workflow)
	}
	requireContains(t, current.Status, "development")
	requireContains(t, current.Status, "production")
	requireSliceContains(t, current.Allowed, "select the configured development, testnet, or production GitHub environment for manual dispatch without accepting live URL inputs")
	requireSliceContains(t, current.Allowed, "reuse an existing immutable ECR image tag by resolving its digest instead of overwriting the tag")
	if len(current.Allowed) < 5 {
		t.Fatalf("current CD allowed actions are underspecified: %#v", current.Allowed)
	}
}
