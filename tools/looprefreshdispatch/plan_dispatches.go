package main

import "sort"

func dispatchesFromWorkflowMap(byWorkflow map[string]workflowScope) []workflowDispatch {
	workflows := make([]string, 0, len(byWorkflow))
	for workflow := range byWorkflow {
		workflows = append(workflows, workflow)
	}
	sort.Strings(workflows)
	out := make([]workflowDispatch, 0, len(workflows))
	for _, workflow := range workflows {
		scope := byWorkflow[workflow]
		loopIDs := sortedKeysInt(scope.loopCounts)
		out = append(out, workflowDispatch{
			WorkflowFile:     workflow,
			VerifiedCommand:  refreshWorkflowCommand(workflow),
			LoopIDs:          loopIDs,
			CommandCount:     commandCount(scope.loopCounts),
			ClaimIDs:         sortedKeys(scope.claimIDs),
			EvidenceChainIDs: sortedKeys(scope.evidenceChainIDs),
		})
	}
	return out
}
