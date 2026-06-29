package main

import "sort"

func dispatchesFromWorkflowMap(byWorkflow map[string]map[string]int) []workflowDispatch {
	workflows := make([]string, 0, len(byWorkflow))
	for workflow := range byWorkflow {
		workflows = append(workflows, workflow)
	}
	sort.Strings(workflows)
	out := make([]workflowDispatch, 0, len(workflows))
	for _, workflow := range workflows {
		loopIDs := sortedKeysInt(byWorkflow[workflow])
		out = append(out, workflowDispatch{
			WorkflowFile:    workflow,
			VerifiedCommand: refreshWorkflowCommand(workflow),
			LoopIDs:         loopIDs,
			CommandCount:    commandCount(byWorkflow[workflow]),
		})
	}
	return out
}
