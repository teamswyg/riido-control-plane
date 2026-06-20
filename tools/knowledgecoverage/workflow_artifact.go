package main

import "strings"

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
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, "actions/upload-artifact") &&
			uploadBlockNamesArtifact(lines, i, artifact) {
			return true
		}
	}
	return false
}

func workflowTextUploadsArtifactStrict(text, artifact string) bool {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, "actions/upload-artifact") &&
			uploadBlockNamesArtifact(lines, i, artifact) &&
			uploadBlockFailsOnMissingFiles(lines, i) {
			return true
		}
	}
	return false
}

func uploadBlockNamesArtifact(lines []string, start int, artifact string) bool {
	for i := start + 1; i < len(lines) && i <= start+20; i++ {
		if nextArtifactUpload(lines[i]) {
			return false
		}
		if workflowNameValue(lines[i]) == artifact {
			return true
		}
	}
	return false
}

func uploadBlockFailsOnMissingFiles(lines []string, start int) bool {
	for i := start + 1; i < len(lines) && i <= start+20; i++ {
		if nextArtifactUpload(lines[i]) {
			return false
		}
		if workflowIfNoFilesFoundValue(lines[i]) == "error" {
			return true
		}
	}
	return false
}

func workflowNameValue(line string) string {
	value, ok := strings.CutPrefix(strings.TrimSpace(line), "name:")
	if !ok {
		return ""
	}
	return strings.Trim(strings.TrimSpace(value), `"'`)
}
