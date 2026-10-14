package main

import "strings"

type workflowStep []string

func workflowStepStart(lines []string, index int) int {
	lineIndent := leadingSpaces(lines[index])
	for i := index; i >= 0; i-- {
		if workflowIsStepStart(lines[i]) && leadingSpaces(lines[i]) <= lineIndent {
			return i
		}
	}
	return index
}

func workflowStepEnd(lines []string, start int) int {
	stepIndent := leadingSpaces(lines[start])
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if leadingSpaces(lines[i]) < stepIndent {
			return i
		}
		if leadingSpaces(lines[i]) == stepIndent && workflowIsStepStart(lines[i]) {
			return i
		}
	}
	return len(lines)
}

func workflowIsStepStart(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "- ")
}
