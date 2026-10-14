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

func workflowUploadsEvidenceOutStrict(root, workflowPath, tool, artifact string) bool {
	paths := workflowEvidenceOutPaths(root, workflowPath, tool)
	if len(paths) == 0 {
		return false
	}
	if pipeline, ok := readRiidoPipeline(root, workflowPath); ok {
		for _, path := range paths {
			if riidoPipelineUploadsStrict(pipeline, artifact, path) {
				return true
			}
		}
		return false
	}
	data, err := readWorkflow(root, workflowPath)
	if err != nil {
		return false
	}
	for _, path := range paths {
		if workflowTextUploadsArtifactPathStrict(string(data), artifact, path) {
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

func workflowTextUploadsArtifactPathStrict(text, artifact, path string) bool {
	for _, step := range workflowUploadArtifactSteps(text) {
		if workflowStepNamesArtifact(step, artifact) &&
			workflowStepUploadsPath(step, path) &&
			workflowStepFailsOnMissingFiles(step) {
			return true
		}
	}
	return false
}
