package main

func refreshPlanNextCommands(
	loop loopRecord,
	claims refreshPlanClaimSet,
) []refreshPlanCommand {
	commands := []refreshPlanCommand{{
		Kind:             "refresh_workflow",
		Command:          refreshCommand(loop.RefreshWorkflow),
		ClaimIDs:         sortedCopy(claims.ClaimIDs),
		EvidenceChainIDs: sortedCopy(claims.EvidenceChainIDs),
	}}
	for _, command := range claims.VerifierCommands {
		scope := claims.CommandScopes[command]
		commands = append(commands, refreshPlanCommand{
			Kind:             "target_verifier",
			Command:          command,
			ClaimIDs:         sortedCopy(scope.ClaimIDs),
			EvidenceChainIDs: sortedCopy(scope.EvidenceChainIDs),
		})
	}
	return commands
}
