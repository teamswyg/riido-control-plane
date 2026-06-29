package main

type loopCoverageDimension struct {
	id              string
	loopTokenLabel  string
	claimTokenLabel string
	loopTokens      func(loopRecord) []string
	claimTokens     func(claimBinding) []string
}

var loopCoverageDimensions = []loopCoverageDimension{
	{
		id:              "observes",
		loopTokenLabel:  "observe token",
		claimTokenLabel: "observe token",
		loopTokens:      func(loop loopRecord) []string { return loop.Observes },
		claimTokens:     func(claim claimBinding) []string { return claim.CoversObserves },
	},
	{
		id:              "verifies",
		loopTokenLabel:  "verify token",
		claimTokenLabel: "verify token",
		loopTokens:      func(loop loopRecord) []string { return loop.Verifies },
		claimTokens:     func(claim claimBinding) []string { return claim.CoversVerifies },
	},
	{
		id:              "fails_when",
		loopTokenLabel:  "fail token",
		claimTokenLabel: "fail token",
		loopTokens:      func(loop loopRecord) []string { return loop.FailsWhen },
		claimTokens:     func(claim claimBinding) []string { return claim.CoversFails },
	},
	{
		id:              "evidence",
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
