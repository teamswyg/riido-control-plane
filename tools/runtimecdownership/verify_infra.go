package main

import "fmt"

func verifyInfraBoundary(m manifest) (int, error) {
	if m.Infra.Repo != "riido-infra" || m.InfraTopology.Repo != "riido-infra" {
		return 0, fmt.Errorf("infra repo ownership drifted")
	}
	if m.InfraTopology.RiidoTask != "RIID-4822" {
		return 0, fmt.Errorf("infra topology task drifted")
	}
	if len(m.Infra.Paths) == 0 || len(m.InfraTopology.RequiredOutput) != 2 {
		return 0, fmt.Errorf("infra handoff is underspecified")
	}
	if len(m.InfraVisibility.MustKnow) == 0 || len(m.InfraVisibility.MustNotFrom) == 0 {
		return 0, fmt.Errorf("infra visibility policy is underspecified")
	}
	if !nonEmpty(m.DependencyDirection.TopDown, m.DependencyDirection.BottomUp) {
		return 0, fmt.Errorf("dependency direction must be explicit")
	}
	return len(m.Infra.Paths) + len(m.InfraTopology.RequiredOutput), nil
}
