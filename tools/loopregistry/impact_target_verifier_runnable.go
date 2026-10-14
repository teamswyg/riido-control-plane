package main

func attachRunnableTargetVerifierCommands(plan *targetVerifierPlan) {
	if plan == nil {
		return
	}
	plan.RunnableCommands = targetVerifierRunnableCommands(plan)
	plan.RunnableCommandCount = len(plan.RunnableCommands)
}

func targetVerifierRunnableCommands(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	if len(plan.FocusedCommands) > 0 {
		return append([]string(nil), plan.FocusedCommands...)
	}
	return append([]string(nil), plan.EntrypointCommands...)
}
