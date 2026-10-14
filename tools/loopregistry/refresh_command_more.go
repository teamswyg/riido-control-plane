package main

func selectedLoop(plan refreshPlan) selectedRefreshLoop {
	return selectedRefreshLoop{
		LoopID:            plan.LoopID,
		EvidenceExpiresAt: plan.EvidenceExpiresAt,
		CommandCount:      len(plan.NextCommands),
	}
}

func selectedCommand(loopID string, command refreshPlanCommand) selectedRefreshCommand {
	return selectedRefreshCommand{
		LoopID:           loopID,
		Kind:             command.Kind,
		Command:          command.Command,
		ClaimIDs:         sortedCopy(command.ClaimIDs),
		EvidenceChainIDs: sortedCopy(command.EvidenceChainIDs),
	}
}
