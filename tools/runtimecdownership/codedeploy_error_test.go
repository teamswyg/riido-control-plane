package main

import "testing"

func TestVerifyCodeDeployModeRejectsDrift(t *testing.T) {
	t.Parallel()
	m := testManifest(t)
	mode := m.OptionalModes[0]
	future := m.Future[0]
	cases := map[string]func(optionalWorkflowMode, futureStrategy) (optionalWorkflowMode, futureStrategy){
		"id": func(mode optionalWorkflowMode, future futureStrategy) (optionalWorkflowMode, futureStrategy) {
			mode.ID = "other"
			return mode, future
		},
		"owner": func(mode optionalWorkflowMode, future futureStrategy) (optionalWorkflowMode, futureStrategy) {
			future.TopologyOwner = "other"
			return mode, future
		},
		"inputs": func(mode optionalWorkflowMode, future futureStrategy) (optionalWorkflowMode, futureStrategy) {
			mode.ActivationInputs = nil
			return mode, future
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			badMode, badFuture := mutate(mode, future)
			if err := verifyCodeDeployMode(badMode, badFuture); err == nil {
				t.Fatalf("verifyCodeDeployMode accepted %s drift", name)
			}
		})
	}
}

func TestVerifyCodeDeployGateRejectsDrift(t *testing.T) {
	t.Parallel()
	gate := testManifest(t).CodeDeployActivationGate
	cases := map[string]func(codeDeployActivationGate) codeDeployActivationGate{
		"task": func(gate codeDeployActivationGate) codeDeployActivationGate {
			gate.RiidoTask = "other"
			return gate
		},
		"rule": func(gate codeDeployActivationGate) codeDeployActivationGate {
			gate.Rule = "ambiguous"
			return gate
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := verifyCodeDeployGate(mutate(gate)); err == nil {
				t.Fatalf("verifyCodeDeployGate accepted %s drift", name)
			}
		})
	}
}
