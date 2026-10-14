package main

import "sort"

func dispatchesFromWorkflowMap(byWorkflow map[string]workflowScope) []workflowDispatch {
	commands := make([]string, 0, len(byWorkflow))
	for command := range byWorkflow {
		commands = append(commands, command)
	}
	sort.Strings(commands)
	out := make([]workflowDispatch, 0, len(commands))
	for _, command := range commands {
		scope := byWorkflow[command]
		loopIDs := sortedKeysInt(scope.loopCounts)
		out = append(out, workflowDispatch{
			WorkflowFile:     scope.workflowFile,
			VerifiedCommand:  scope.verifiedCommand,
			Inputs:           scope.inputs,
			LoopIDs:          loopIDs,
			CommandCount:     commandCount(scope.loopCounts),
			ClaimIDs:         sortedKeys(scope.claimIDs),
			EvidenceChainIDs: sortedKeys(scope.evidenceChainIDs),
		})
	}
	return out
}
