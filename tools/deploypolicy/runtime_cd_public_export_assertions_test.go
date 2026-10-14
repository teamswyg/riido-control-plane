package deploypolicy

import "testing"

func assertRuntimeCDPublicExport(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	assertPublicExportContract(t, p.PublicExport)
	assertPublicSurfaceScanContract(t, p.PublicSurfaceScan)
}

func assertPublicExportContract(t *testing.T, contract publicExportContract) {
	t.Helper()
	if contract.RiidoTask != "RIID-4835" {
		t.Fatalf("public export work unit drifted: %#v", contract)
	}
	if contract.CanonicalOwner != "riido-control-plane" || contract.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public export ownership drifted: %#v", contract)
	}
	requireSliceContains(t, contract.AllowedPublicExports, "stable GitHub secret and variable key names")
	requireSliceContains(t, contract.AllowedPublicExports, "stable infra output key names that operators map into GitHub environment variables")
	requireSliceContains(t, contract.AllowedPublicExports, "aggregate pass or fail status without live payload values")
	requireSliceContains(t, contract.AllowedPublicExports, "redacted live workflow summary artifact names and paths")
	requireSliceContains(t, contract.ForbiddenPublicExports, "image URIs or digests")
	requireSliceContains(t, contract.ForbiddenPublicExports, "ECS task-definition JSON")
	requireSliceContains(t, contract.ForbiddenPublicExports, "CodeDeploy create-deployment request JSON")
	requireSliceContains(t, contract.ForbiddenPublicExports, "Terraform plan output, state, tfvars, apply logs, or raw operator evidence")
	requireSliceContains(t, contract.InfraMustConsumeOnly, "stable output names")
	requireSliceContains(t, contract.InfraMustConsumeOnly, "redaction categories")
	requireSliceContains(t, contract.InfraMustConsumeOnly, "operator evidence summaries stored outside public repositories")
	requireSliceContains(t, contract.WorkflowMustNotUse, "actions/upload-artifact for live deployment payloads")
	requireSliceContains(t, contract.WorkflowMustNotUse, "GITHUB_OUTPUT for live deployment values")
	requireSliceContains(t, contract.WorkflowMustNotUse, "workflow_dispatch inputs for live URLs")
}

func assertPublicSurfaceScanContract(t *testing.T, contract publicSurfaceScanContract) {
	t.Helper()
	if contract.RiidoTask != "RIID-4836" {
		t.Fatalf("public surface scan work unit drifted: %#v", contract)
	}
	if contract.CanonicalOwner != "riido-control-plane" || contract.InfraAwarenessOwner != "riido-infra" {
		t.Fatalf("public surface scan ownership drifted: %#v", contract)
	}
	requireSliceContains(t, contract.ScopePaths, ".github/workflows/deploy-ai-agent-testnet.yml")
	requireSliceContains(t, contract.ScopePaths, "docs/30-architecture/api-client-delivery.md")
	requireSliceContains(t, contract.ScopePaths, "docs/30-architecture/runtime-cd-ownership.md")
	requireSliceContains(t, contract.ScopePaths, "web/generated/aiAgentClient.react.ts")
	requireSliceContains(t, contract.ForbiddenLiterals, "ai-api.riido.io")
	requireSliceContains(t, contract.WorkflowForbiddenMechanism, "GITHUB_OUTPUT")
	requireSliceContains(t, contract.AllowedPublicSurface, "AWS CLI response field names inside the deploy workflow")
	requireSliceContains(t, contract.AllowedPublicSurface, "redacted live workflow summary artifact names and paths")
	requireContains(t, contract.InfraMustTreatScanAs, "awareness policy only")
}
