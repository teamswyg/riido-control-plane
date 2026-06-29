package main

type workflowScope struct {
	workflowFile     string
	verifiedCommand  string
	inputs           []workflowInput
	loopCounts       map[string]int
	claimIDs         map[string]bool
	evidenceChainIDs map[string]bool
}

func newWorkflowScope(workflowFile, verifiedCommand string, inputs []workflowInput) workflowScope {
	return workflowScope{
		workflowFile:     workflowFile,
		verifiedCommand:  verifiedCommand,
		inputs:           inputs,
		loopCounts:       map[string]int{},
		claimIDs:         map[string]bool{},
		evidenceChainIDs: map[string]bool{},
	}
}

func (scope workflowScope) add(command selectedRefreshCommand) {
	scope.loopCounts[command.LoopID]++
	for _, id := range command.ClaimIDs {
		scope.claimIDs[id] = true
	}
	for _, id := range command.EvidenceChainIDs {
		scope.evidenceChainIDs[id] = true
	}
}
