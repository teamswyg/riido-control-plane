package deploypolicy

import "testing"

func assertRuntimeCDConfigKeys(t *testing.T, p runtimeCDOwnership, f runtimeCDDocFixture) {
	t.Helper()
	config := p.PublicConfigKeyMinimization
	if config.RiidoTask != "RIID-4839" {
		t.Fatalf("public config key minimization work unit drifted: %#v", config)
	}
	if config.CanonicalOwner != "riido-control-plane" || config.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public config key minimization ownership drifted: %#v", config)
	}
	requireContains(t, config.Rule, "minimum stable GitHub configuration keys")
	requireContains(t, config.WorkflowKeySource, "any additional RIIDO_AI_SERVER_*")
	requireStringSetExact(t, config.RequiredSecretKeys, expectedSecretKeys())
	requireStringSetExact(t, config.RequiredVariableKeys, expectedRequiredVars())
	requireStringSetExact(t, config.OptionalVariableKeys, expectedOptionalVars())
	assertStableInfraSourceNames(t, config.StableInfraSourceNames)
	assertPublicDocKeyPolicy(t, config)
	workflowRefs := f.DeployWorkflow + "\n" + f.SmokeWorkflow
	requireStringSetExact(t, collectGitHubConfigRefs(workflowRefs, "secrets"), expectedSecretKeys())
	requireStringSetExact(t, collectGitHubConfigRefs(workflowRefs, "vars"), expectedAllVars())
}

func assertStableInfraSourceNames(t *testing.T, names []string) {
	t.Helper()
	for _, name := range []string{"ecr_repository_name", "ecs_cluster_name", "service_name", "container_name", "codedeploy_application_name", "codedeploy_deployment_group_name"} {
		requireSliceContains(t, names, name)
	}
}

func assertPublicDocKeyPolicy(t *testing.T, config publicConfigKeyMinimization) {
	t.Helper()
	requireSliceContains(t, config.PublicDocsMayReference, "required or optional key categories")
	requireSliceContains(t, config.PublicDocsMayReference, "the machine-readable manifest path that contains the exact key list")
	requireSliceContains(t, config.PublicDocsMustNotRef, "exact deploy or smoke key-name lists outside the machine-readable manifest and workflow files")
	requireSliceContains(t, config.PublicDocsMustNotRef, "environment-specific example values")
	requireSliceContains(t, config.PublicDocsMustNotRef, "live hosts")
}
