package main

func verifyLoopEvidenceCoverage(m manifest) error {
	return verifyLoopCoverageDimension(m, loopCoverageDimensionByID("evidence"))
}
