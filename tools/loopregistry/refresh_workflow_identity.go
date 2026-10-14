package main

import "strings"

func refreshWorkflowDeclaresLoopID(text, loopID string) bool {
	if strings.TrimSpace(loopID) == "" {
		return false
	}
	if !strings.Contains(text, "RIIDO_LOOP_IDS") {
		return false
	}
	return workflowLoopIDs(text)[loopID]
}

func workflowLoopIDs(text string) map[string]bool {
	ids := map[string]bool{}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "RIIDO_LOOP_IDS") {
			continue
		}
		for _, field := range strings.Fields(line) {
			ids[cleanWorkflowLoopIDField(field)] = true
		}
	}
	delete(ids, "")
	delete(ids, "RIIDO_LOOP_IDS")
	return ids
}

func cleanWorkflowLoopIDField(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	value = strings.TrimPrefix(value, "RIIDO_LOOP_IDS:")
	return strings.Trim(value, `"'`)
}
