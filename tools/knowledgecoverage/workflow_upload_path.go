package main

func workflowUploadsEvidenceOut(root, workflowPath, tool, artifact string) bool {
	paths := workflowEvidenceOutPaths(root, workflowPath, tool)
	if len(paths) == 0 {
		return false
	}
	data, err := readWorkflow(root, workflowPath)
	if err != nil {
		return false
	}
	for _, path := range paths {
		if workflowTextUploadsArtifactPath(string(data), artifact, path) {
			return true
		}
	}
	return false
}

func workflowTextUploadsArtifactPath(text, artifact, path string) bool {
	for _, step := range workflowUploadArtifactSteps(text) {
		if workflowStepNamesArtifact(step, artifact) &&
			workflowStepUploadsPath(step, path) {
			return true
		}
	}
	return false
}
