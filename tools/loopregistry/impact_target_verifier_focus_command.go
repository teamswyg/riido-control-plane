package main

func focusedVerifierCommands(
	claimIDs []string,
	surfaces []claimSurface,
	units []targetVerifierCommand,
) []string {
	if len(claimIDs) == 0 {
		return nil
	}
	if len(surfaces) > 0 {
		return focusedVerifierCommandsFromSurfaces(claimIDs, surfaces)
	}
	return focusedVerifierCommandsFromUnits(claimIDs, units)
}

func focusedVerifierCommandsFromUnits(
	claimIDs []string,
	units []targetVerifierCommand,
) []string {
	claims := stringSet(claimIDs)
	out := []string{}
	for _, unit := range units {
		if targetVerifierCommandHasClaim(unit, claims) {
			out = appendUnique(out, unit.Command)
		}
	}
	return out
}

func focusedVerifierCommandsFromSurfaces(
	claimIDs []string,
	surfaces []claimSurface,
) []string {
	claims := stringSet(claimIDs)
	out := []string{}
	for _, surface := range surfaces {
		if claims[surface.ID] {
			out = appendUnique(out, surface.VerifierCommands...)
		}
	}
	return out
}

func targetVerifierCommandHasClaim(
	unit targetVerifierCommand,
	claims map[string]bool,
) bool {
	for _, claimID := range unit.ClaimIDs {
		if claims[claimID] {
			return true
		}
	}
	return false
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[value] = true
	}
	return out
}
