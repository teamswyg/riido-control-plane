package main

import "testing"

func TestLoopRegistryEvidenceExposesCoverageDimensions(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if len(got.CoverageDimensions) != len(loopCoverageDimensions) {
		t.Fatalf("coverage dimensions = %d, want %d",
			len(got.CoverageDimensions), len(loopCoverageDimensions))
	}
	dim := coverageDimensionByID(got.CoverageDimensions, "evidence")
	if dim.ID == "" || dim.LoopField == "" || dim.ClaimField == "" ||
		dim.LoopTokenLabel == "" || dim.ClaimTokenLabel == "" {
		t.Fatalf("incomplete coverage dimension: %+v", dim)
	}
}

func coverageDimensionByID(
	dimensions []loopCoverageDimensionSurface,
	id string,
) loopCoverageDimensionSurface {
	for _, dim := range dimensions {
		if dim.ID == id {
			return dim
		}
	}
	return loopCoverageDimensionSurface{}
}
