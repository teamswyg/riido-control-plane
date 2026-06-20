package main

func workflowUploadsArtifact(root, workflowPath, artifact string) bool {
	data, err := readWorkflow(root, workflowPath)
	if err != nil {
		return false
	}
	return workflowTextUploadsArtifact(string(data), artifact)
}

func workflowUploadsArtifactStrict(root, workflowPath, artifact string) bool {
	data, err := readWorkflow(root, workflowPath)
	if err != nil {
		return false
	}
	return workflowTextUploadsArtifactStrict(string(data), artifact)
}

func workflowTextUploadsArtifact(text, artifact string) bool {
	for _, step := range workflowUploadArtifactSteps(text) {
		if workflowStepNamesArtifact(step, artifact) {
			return true
		}
	}
	return false
}

func workflowTextUploadsArtifactStrict(text, artifact string) bool {
	for _, step := range workflowUploadArtifactSteps(text) {
		if workflowStepNamesArtifact(step, artifact) &&
			workflowStepFailsOnMissingFiles(step) {
			return true
		}
	}
	return false
}
