package main

func missingCoverageDimensions(claim registryClaim, loop registryLoop) []string {
	missing := []string{}
	if len(loop.Observes) > 0 && len(claim.CoversObserves) == 0 {
		missing = append(missing, "covers_observes")
	}
	if len(loop.Verifies) > 0 && len(claim.CoversVerifies) == 0 {
		missing = append(missing, "covers_verifies")
	}
	if len(loop.FailsWhen) > 0 && len(claim.CoversFails) == 0 {
		missing = append(missing, "covers_fails_when")
	}
	return missing
}
