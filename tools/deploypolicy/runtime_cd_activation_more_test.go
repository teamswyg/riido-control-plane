package deploypolicy

import "testing"

func assertCodeDeployActivationGateDetails(t *testing.T, gate codeDeployActivationGate) {
	t.Helper()
	requireSliceContains(t, gate.PublicRepoMustNotKeep, "environment-specific CodeDeploy application or deployment-group values")
	requireSliceContains(t, gate.PublicRepoMustNotKeep, "generated CodeDeploy AppSpec or request JSON")
	requireSliceContains(t, gate.PublicRepoMustNotKeep, "workflow run URL as deploy evidence")
	requireSliceContains(t, gate.PublicRepoMustNotKeep, "Terraform plan, state, tfvars, apply logs, or raw operator evidence")
	requireSliceContains(t, gate.InfraMustKnow, "activation does not move create/wait/smoke execution out of riido-control-plane")
	requireSliceContains(t, gate.InfraMustKnow, "public repos should not request convenience handoff payloads from infra or the deploy workflow")
}
