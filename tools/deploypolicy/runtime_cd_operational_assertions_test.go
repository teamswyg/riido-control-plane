package deploypolicy

import "testing"

func assertRuntimeCDOperationalDetail(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	assertOperationalMinimization(t, p.PublicOperationalDetailMinimization)
	assertCodeDeployActivationGate(t, p.CodeDeployActivationGate)
	assertCodeDeployActivationGateDetails(t, p.CodeDeployActivationGate)
	assertBroadDocsDoNotListCDKeys(t, p)
}

func assertOperationalMinimization(t *testing.T, policy operationalDetailMinimization) {
	t.Helper()
	if policy.RiidoTask != "RIID-4853" {
		t.Fatalf("public operational detail minimization work unit drifted: %#v", policy)
	}
	if policy.CanonicalOwner != "riido-control-plane" || policy.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public operational detail minimization ownership drifted: %#v", policy)
	}
	requireContains(t, policy.Rule, "smallest useful CD description")
	requireContains(t, policy.Rule, "Stable non-secret operational details")
	requireSliceContains(t, policy.PublicRepoMayKeep, "workflow names and trigger policy")
	requireSliceContains(t, policy.PublicRepoMayKeep, "stable infra source names without values")
	requireSliceContains(t, policy.PublicRepoShouldAvoid, "exact deploy or smoke key-name lists outside the machine-readable manifest and workflow files")
	requireSliceContains(t, policy.PublicRepoShouldAvoid, "duplicating operational setup details in broad README, client-facing docs, generated-client docs, or PR prose when a link to the manifest is sufficient")
	requireSliceContains(t, policy.InfraMustKnow, "the CD ownership remodel is settled: runtime artifact CD remains in riido-control-plane")
	requireSliceContains(t, policy.InfraMustKnow, "tightening public operational disclosure is Terraform no-diff unless a future SSOT asks for topology, secret, IAM, network, persistence, or evidence tooling changes")
}

func assertCodeDeployActivationGate(t *testing.T, gate codeDeployActivationGate) {
	t.Helper()
	if gate.RiidoTask != "RIID-4855" || gate.Status != "topology-ready-operator-environment-gated" {
		t.Fatalf("CodeDeploy activation gate drifted: %#v", gate)
	}
	if gate.CanonicalOwner != "riido-control-plane" || gate.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("CodeDeploy activation gate ownership drifted: %#v", gate)
	}
	requireContains(t, gate.Rule, "not an infra-owned deployment action")
	requireContains(t, gate.Rule, "operators map the infra-provided application/deployment-group names")
	requireSliceContains(t, gate.ActivationRequirements, "riido-infra CodeDeploy topology work unit has been applied and reviewed outside public repositories")
	requireSliceContains(t, gate.ActivationRequirements, "both optional CodeDeploy GitHub environment variables are configured together")
	requireSliceContains(t, gate.ActivationRequirements, "the deploy workflow creates and waits for the CodeDeploy deployment in the same job")
	requireSliceContains(t, gate.PublicRepoMayKeep, "stable activation key categories")
	requireSliceContains(t, gate.PublicRepoMayKeep, "aggregate activation readiness or pass-fail status without live payload values")
}
