package main

type loopCoverageDimension struct {
	id              string
	loopField       string
	claimField      string
	loopTokenLabel  string
	claimTokenLabel string
	loopTokens      func(loopRecord) []string
	claimTokens     func(claimBinding) []string
}

var loopCoverageDimensions = []loopCoverageDimension{
	{
		id:              "observes",
		loopField:       "loops[].observes",
		claimField:      "claim_bindings[].covers_observes",
		loopTokenLabel:  "observe token",
		claimTokenLabel: "observe token",
		loopTokens:      func(loop loopRecord) []string { return loop.Observes },
		claimTokens:     func(claim claimBinding) []string { return claim.CoversObserves },
	},
	{
		id:              "verifies",
		loopField:       "loops[].verifies",
		claimField:      "claim_bindings[].covers_verifies",
		loopTokenLabel:  "verify token",
		claimTokenLabel: "verify token",
		loopTokens:      func(loop loopRecord) []string { return loop.Verifies },
		claimTokens:     func(claim claimBinding) []string { return claim.CoversVerifies },
	},
	{
		id:              "fails_when",
		loopField:       "loops[].fails_when",
		claimField:      "claim_bindings[].covers_fails_when",
		loopTokenLabel:  "fail token",
		claimTokenLabel: "fail token",
		loopTokens:      func(loop loopRecord) []string { return loop.FailsWhen },
		claimTokens:     func(claim claimBinding) []string { return claim.CoversFails },
	},
	{
		id:              "evidence",
		loopField:       "loops[].evidence[].path",
		claimField:      "claim_bindings[].covers_evidence",
		loopTokenLabel:  "evidence source",
		claimTokenLabel: "evidence source",
		loopTokens:      loopEvidenceSourcePaths,
		claimTokens:     func(claim claimBinding) []string { return claim.CoversEvidence },
	},
}

func loopCoverageDimensionByID(id string) loopCoverageDimension {
	for _, dim := range loopCoverageDimensions {
		if dim.id == id {
			return dim
		}
	}
	return loopCoverageDimension{}
}
