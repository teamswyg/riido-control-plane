package main

type refreshCommandScope struct {
	ClaimIDs         []string
	EvidenceChainIDs []string
}

func ensureRefreshCommandScopes(
	scopes map[string]refreshCommandScope,
) map[string]refreshCommandScope {
	if scopes != nil {
		return scopes
	}
	return map[string]refreshCommandScope{}
}

func addRefreshCommandScopes(
	scopes map[string]refreshCommandScope,
	claimID string,
	surface claimSurface,
) {
	for _, command := range surface.VerifierCommands {
		scope := scopes[command]
		scope.ClaimIDs = appendMissingStrings(scope.ClaimIDs, []string{claimID})
		scope.EvidenceChainIDs = appendMissingStrings(
			scope.EvidenceChainIDs,
			surface.EvidenceChainIDs,
		)
		scopes[command] = scope
	}
}
