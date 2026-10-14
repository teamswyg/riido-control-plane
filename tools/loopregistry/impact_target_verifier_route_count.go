package main

func attachTargetVerifierRouteCounts(plan *targetVerifierPlan) {
	direct, fallback := targetVerifierRouteCounts(plan.RoutingPackets)
	plan.RoutingDirectCount = direct
	plan.RoutingFallbackCount = fallback
}

func targetVerifierRouteCounts(
	routes []targetVerifierRoute,
) (int, int) {
	direct := 0
	fallback := 0
	for _, route := range routes {
		if route.UsesEntrypointFallback {
			fallback++
			continue
		}
		direct++
	}
	return direct, fallback
}
