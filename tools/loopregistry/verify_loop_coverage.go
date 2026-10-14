package main

func verifyLoopVerifyCoverage(m manifest) error {
	return verifyLoopCoverageDimension(m, loopCoverageDimensionByID("verifies"))
}
