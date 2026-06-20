package deploypolicy

import "testing"

func assertRuntimeCDRedactionAndHandoff(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	assertRuntimeCDRedactionPolicy(t, p.Redaction)
	assertRuntimeCDHandoffPolicy(t, p.Handoff)
}

func assertRuntimeCDRedactionPolicy(t *testing.T, policy redactionPolicy) {
	t.Helper()
	for _, forbidden := range forbiddenRuntimeCDPublicValues() {
		requireSliceContains(t, policy.MustNot, forbidden)
	}
	requireSliceContains(t, policy.Workflows, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, policy.Workflows, ".github/workflows/ai-agent-client-testnet-smoke.yml")
	requireSliceContains(t, policy.MayCommit, "workflow file")
	requireSliceContains(t, policy.MayCommit, "non-live behavior documentation")
	requireSliceContains(t, policy.ShouldMinimize, "publish only stable configuration key names that operators must set")
	requireSliceContains(t, policy.ShouldMinimize, "centralize required deploy key names in the workflow files and machine-readable ownership manifest instead of scattering environment-specific examples")
	requireSliceContains(t, policy.ShouldMinimize, "keep exact deploy key-name lists out of human-readable public docs except the workflow files that consume them")
	requireSliceContains(t, policy.ShouldMinimize, "avoid environment-specific examples for domains, clusters, services, applications, deployment groups, ARNs, and URLs")
	requireSliceContains(t, policy.ShouldMinimize, "apply restrictive file permissions before writing live handoff, task-definition, CodeDeploy, or smoke replay files")
}

func assertRuntimeCDHandoffPolicy(t *testing.T, policy handoffPolicy) {
	t.Helper()
	if policy.Scope != "same-job-runner-temp-only" {
		t.Fatalf("handoff scope drifted: %#v", policy)
	}
	requireSliceContains(t, policy.AllowedStorage, "$RUNNER_TEMP files created under umask 077 and chmod 600 before reuse")
	requireSliceContains(t, policy.RequiredCleanup, "remove image URI, ECS task-definition ARN, and container-port temp files in an always-running cleanup step")
	requireSliceContains(t, policy.RequiredCleanup, "remove generated CodeDeploy AppSpec, request JSON, and deployment-id files with same-step traps")
	requireSliceContains(t, policy.RequiredCleanup, "remove smoke replay temp files with same-step traps")
	requireSliceContains(t, policy.ForbiddenMechanism, "GitHub step outputs for live deployment values")
	requireSliceContains(t, policy.ForbiddenMechanism, "uploaded workflow artifacts")
}
