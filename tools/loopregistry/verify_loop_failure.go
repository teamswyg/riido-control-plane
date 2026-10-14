package main

func verifyLoopFailureCoverage(m manifest) error {
	return verifyLoopCoverageDimension(m, loopCoverageDimensionByID("fails_when"))
}
