package deploypolicy

import "testing"

func assertRuntimeCDStrategies(t *testing.T, p runtimeCDOwnership) {
	t.Helper()
	assertRuntimeCDCurrentStrategy(t, p.Current)
	assertOptionalCodeDeployMode(t, p.OptionalModes)
	assertFutureCodeDeployStrategy(t, p.Future)
}

func assertOptionalCodeDeployMode(t *testing.T, modes []optionalWorkflowMode) {
	t.Helper()
	for _, mode := range modes {
		if mode.ID != "codedeploy-blue-green" {
			continue
		}
		if mode.CDOwner != "riido-control-plane" || mode.TopologyOwner != "riido-infra" {
			t.Fatalf("optional CodeDeploy owner drifted: %#v", mode)
		}
		requireSliceContains(t, mode.ActivationInputs, "RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION")
		requireSliceContains(t, mode.ActivationInputs, "RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP")
		requireSliceContains(t, mode.Allowed, "create CodeDeploy AppSpec content in same-job runner temp files")
		requireSliceContains(t, mode.Allowed, "wait for CodeDeploy deployment success")
		requireSliceContains(t, mode.MustNotOwn, "CodeDeploy application or deployment group topology")
		return
	}
	t.Fatal("optional CodeDeploy workflow mode is missing")
}

func assertFutureCodeDeployStrategy(t *testing.T, future []futureStrategy) {
	t.Helper()
	for _, item := range future {
		if item.ID != "codedeploy-blue-green" {
			continue
		}
		if item.CDOwner != "riido-control-plane" || item.TopologyOwner != "riido-infra" {
			t.Fatalf("CodeDeploy owner drifted: %#v", item)
		}
		requireSliceContains(t, item.ControlPlaneMayOwn, "create CodeDeploy deployment from the same-job immutable image value and infra-provided deployment target")
		requireSliceContains(t, item.ControlPlaneMayOwn, "wait for CodeDeploy deployment completion")
		requireSliceContains(t, item.InfraMustOwn, "CodeDeploy application and deployment group")
		requireSliceContains(t, item.InfraMustOwn, "blue green target groups and listener topology")
		return
	}
	t.Fatal("future CodeDeploy strategy is missing")
}
