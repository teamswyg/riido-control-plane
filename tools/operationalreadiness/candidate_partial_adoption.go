package main

func partialRequiredArtifacts(source producerSource, partial partialCheck) []string {
	artifacts := []string{partial.NextArtifact}
	return append(artifacts, source.RequiredNextArtifacts...)
}

func partialAdoptionPlan(source producerSource, partial partialCheck) []adoptionStep {
	steps := []adoptionStep{
		{Artifact: partial.NextArtifact, Command: partial.NextCommand},
	}
	return append(steps, readinessAdoptionPlan(source)...)
}
