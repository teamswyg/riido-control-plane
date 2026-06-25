package main

import "strings"

const sourceCoverageSeedSuffix = ":source_coverage_seed"

func isSourceCoverageSeed(decision decisionRecord) bool {
	return strings.HasSuffix(decision.CandidateID, sourceCoverageSeedSuffix)
}
