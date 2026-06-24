package main

import "fmt"

func verifyAdoptionPlan(item closedLoopCandidate) error {
	byArtifact := map[string]string{}
	for _, step := range item.AdoptionPlan {
		if step.Artifact == "" || step.Command == "" {
			return fmt.Errorf("candidate %s adoption plan step must bind artifact and command", item.ID)
		}
		if byArtifact[step.Artifact] != "" {
			return fmt.Errorf("candidate %s duplicates adoption artifact %s", item.ID, step.Artifact)
		}
		byArtifact[step.Artifact] = step.Command
	}
	for _, artifact := range item.RequiredNextArtifacts {
		if byArtifact[artifact] == "" {
			return fmt.Errorf("candidate %s adoption plan missing artifact %s", item.ID, artifact)
		}
	}
	if len(byArtifact) != len(item.RequiredNextArtifacts) {
		return fmt.Errorf("candidate %s adoption plan has extra artifacts", item.ID)
	}
	return nil
}
