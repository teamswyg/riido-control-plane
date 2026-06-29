package main

func verifyLoopObservationCoverage(m manifest) error {
	return verifyLoopCoverageDimension(m, loopCoverageDimensionByID("observes"))
}
