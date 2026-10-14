package main

import "fmt"

func targetVerifierRouteSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	if plan == nil || plan.RoutingPacketCount == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	return fmt.Sprintf(
		"routes: %d packets, %d direct, %d fallback in %s",
		plan.RoutingPacketCount,
		plan.RoutingDirectCount,
		plan.RoutingFallbackCount,
		evidenceRef,
	)
}
