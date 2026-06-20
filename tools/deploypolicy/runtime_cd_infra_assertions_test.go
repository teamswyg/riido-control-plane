package deploypolicy

import "testing"

func assertRuntimeCDInfra(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	assertInfraConsumes(t, p.Infra)
	assertInfraTopology(t, p.InfraTopology)
	assertInfraVisibility(t, p.InfraVisibility)
}

func assertInfraConsumes(t *testing.T, infra infraConsumes) {
	t.Helper()
	if infra.Repo != "riido-infra" {
		t.Fatalf("infra consumer repo drifted: %q", infra.Repo)
	}
	for _, path := range requiredInfraAwarenessPaths() {
		requireSliceContains(t, infra.Paths, path)
	}
}

func assertInfraTopology(t *testing.T, topology infraTopologyContract) {
	t.Helper()
	if topology.RiidoTask != "RIID-4822" || topology.Repo != "riido-infra" {
		t.Fatalf("infra topology contract drifted: %#v", topology)
	}
	requireSliceContains(t, topology.RequiredOutput, "codedeploy_application_name")
	requireSliceContains(t, topology.RequiredOutput, "codedeploy_deployment_group_name")
	requireSliceContains(t, topology.MustNotConsume, "CodeDeploy service role ARN")
	requireSliceContains(t, topology.MustNotConsume, "target group ARN")
}

func assertInfraVisibility(t *testing.T, policy infraVisibilityPolicy) {
	t.Helper()
	if policy.Repo != "riido-infra" {
		t.Fatalf("infra visibility repo drifted: %q", policy.Repo)
	}
	requireSliceContains(t, policy.MustKnow, "riido-control-plane owns runtime artifact CD execution")
	requireSliceContains(t, policy.MustKnow, "infra-local awareness guards verify the no-diff boundary without consuming public workflow live payloads")
	requireSliceContains(t, policy.MustNotFrom, "generated CodeDeploy AppSpec JSON")
	requireSliceContains(t, policy.MustNotFrom, "image digests or image URIs")
	requireSliceContains(t, policy.MustNotFrom, "smoke replay temp files")
}
