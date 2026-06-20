package main

import "fmt"

func verifyStrategies(root string, m manifest) (int, error) {
	if err := verifyCurrentStrategy(root, m.Current); err != nil {
		return 0, err
	}
	if len(m.OptionalModes) != 1 || len(m.Future) != 1 {
		return 0, fmt.Errorf("expected one optional and one future strategy")
	}
	if err := verifyCodeDeployMode(m.OptionalModes[0], m.Future[0]); err != nil {
		return 0, err
	}
	if err := verifyCodeDeployGate(m.CodeDeployActivationGate); err != nil {
		return 0, err
	}
	return 3, nil
}

func verifyCurrentStrategy(root string, strategy currentStrategy) error {
	if strategy.ID != "ecs-rolling-task-definition-revision" {
		return fmt.Errorf("unexpected current strategy %q", strategy.ID)
	}
	if strategy.CDOwner != "riido-control-plane" || strategy.TopologyOwner != "riido-infra" {
		return fmt.Errorf("current strategy ownership drifted")
	}
	if !pathExists(repoPath(root, strategy.Workflow)) {
		return fmt.Errorf("current strategy workflow does not exist: %s", strategy.Workflow)
	}
	if !containsText(strategy.Allowed, "profile thumbnail upload-intent smoke") {
		return fmt.Errorf("deploy smoke must cover profile thumbnail upload-intent")
	}
	return nil
}

func verifyCodeDeployMode(mode optionalWorkflowMode, future futureStrategy) error {
	if mode.ID != "codedeploy-blue-green" || future.ID != "codedeploy-blue-green" {
		return fmt.Errorf("CodeDeploy strategy id drifted")
	}
	if mode.CDOwner != "riido-control-plane" || future.TopologyOwner != "riido-infra" {
		return fmt.Errorf("CodeDeploy ownership drifted")
	}
	if len(mode.ActivationInputs) != 2 || len(mode.MustNotOwn) == 0 || len(future.InfraMustOwn) == 0 {
		return fmt.Errorf("CodeDeploy gate is underspecified")
	}
	return nil
}

func verifyCodeDeployGate(gate codeDeployActivationGate) error {
	if gate.RiidoTask != "RIID-4855" || gate.Status != "topology-ready-operator-environment-gated" {
		return fmt.Errorf("CodeDeploy activation gate drifted")
	}
	if !containsText([]string{gate.Rule}, "not an infra-owned deployment action") {
		return fmt.Errorf("CodeDeploy gate must preserve ownership rule")
	}
	return nil
}
