package main

import "strings"

func workflowStepBlocks(text string) []string {
	var blocks []string
	var current []string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "      - ") {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
			}
			current = []string{line}
			continue
		}
		if len(current) > 0 {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func workflowStepRunsAlways(block string) bool {
	return strings.Contains(block, "if: always()")
}

func workflowHasAlwaysStep(text string, required ...string) bool {
	for _, block := range workflowStepBlocks(text) {
		if !workflowStepRunsAlways(block) || !containsAll(block, required) {
			continue
		}
		return true
	}
	return false
}

func containsAll(text string, values []string) bool {
	for _, value := range values {
		if !strings.Contains(text, value) {
			return false
		}
	}
	return true
}
