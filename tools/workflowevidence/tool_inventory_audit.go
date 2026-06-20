package main

func auditEvidenceTools(root string, workflowTexts []string) (int, int, []string) {
	tools := evidenceToolDirs(root)
	covered := 0
	var missing []string
	for _, tool := range tools {
		if workflowCallsEvidenceTool(workflowTexts, tool) {
			covered++
			continue
		}
		missing = append(missing, tool)
	}
	return len(tools), covered, uniqueStrings(missing)
}
