package main

func harnessWorkflowProducesCandidates(text, artifact string) bool {
	return workflowHasAlwaysStep(
		text,
		"go run",
		"-candidate-out",
		"out/"+artifact+".json",
	) && workflowAlwaysUploadsStrictArtifact(text, artifact)
}
