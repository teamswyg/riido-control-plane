package main

func refreshPlanNextCommands(
	loop loopRecord,
	claims refreshPlanClaimSet,
) []refreshPlanCommand {
	commands := []refreshPlanCommand{{
		Kind:    "refresh_workflow",
		Command: refreshCommand(loop.RefreshWorkflow),
	}}
	for _, command := range claims.VerifierCommands {
		commands = append(commands, refreshPlanCommand{
			Kind:    "target_verifier",
			Command: command,
		})
	}
	return commands
}
