package main

import "strings"

func workflowStepUploadsPath(step workflowStep, path string) bool {
	for _, value := range workflowStepPathValues(step) {
		if value == path {
			return true
		}
	}
	return false
}

func workflowStepPathValues(step workflowStep) []string {
	var values []string
	for i := 0; i < len(step); i++ {
		value, ok := workflowPathValue(step[i])
		if !ok {
			continue
		}
		if workflowRunValueIsBlock(value) {
			values = append(values, workflowPathBlockValues(step, i+1, leadingSpaces(step[i]))...)
			continue
		}
		values = append(values, workflowCleanCommandValue(value))
	}
	return emptyIfNil(values)
}

func workflowPathValue(line string) (string, bool) {
	value, ok := strings.CutPrefix(strings.TrimSpace(line), "path:")
	if !ok {
		return "", false
	}
	return strings.TrimSpace(value), true
}

func workflowPathBlockValues(step workflowStep, start, pathIndent int) []string {
	var values []string
	for i := start; i < len(step); i++ {
		if strings.TrimSpace(step[i]) == "" {
			continue
		}
		if leadingSpaces(step[i]) <= pathIndent {
			break
		}
		values = append(values, workflowCleanCommandValue(step[i]))
	}
	return values
}
