package main

type workflowScope struct {
	loopCounts       map[string]int
	claimIDs         map[string]bool
	evidenceChainIDs map[string]bool
}

func newWorkflowScope() workflowScope {
	return workflowScope{
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
