package main

func attachTargetVerifierRoutingPackets(plan *targetVerifierPlan) {
	if plan == nil {
		return
	}
	plan.RoutingPackets = targetVerifierRoutingPackets(plan)
}

func targetVerifierRoutingPackets(
	plan *targetVerifierPlan,
) []targetVerifierRoute {
	if plan == nil {
		return nil
	}
	out := make([]targetVerifierRoute, 0, len(plan.Paths))
	for _, path := range plan.Paths {
		commands, direct := targetVerifierRouteCommands(
			path,
			plan.RunnableCommands,
		)
		out = append(out, targetVerifierRoute{
			Path:                   path.Path,
			Component:              path.Component,
			MatchKind:              path.MatchKind,
			RunnableCommands:       commands,
			RunnableCommandCount:   len(commands),
			DirectCommandCount:     direct,
			UsesEntrypointFallback: direct == 0 && len(commands) > 0,
			LoopIDs:                sortedCopy(path.LoopIDs),
			ClaimIDs:               sortedCopy(path.ClaimIDs),
			EvidenceChainIDs:       sortedCopy(path.EvidenceChainIDs),
		})
	}
	return out
}

func targetVerifierRouteCommands(
	path targetVerifierPath,
	runnable []string,
) ([]string, int) {
	direct := intersectStrings(runnable, path.VerifierCommands)
	if len(direct) > 0 {
		return direct, len(direct)
	}
	return sortedCopy(runnable), 0
}

func intersectStrings(left, right []string) []string {
	allowed := map[string]bool{}
	for _, value := range right {
		allowed[value] = true
	}
	var out []string
	for _, value := range left {
		if allowed[value] {
			out = appendUnique(out, value)
		}
	}
	return out
}
