package main

const (
	pressureCandidateSchema   = "riido-control-plane-closed-loop-candidates.v1"
	pressureCandidateSourceID = "control-plane-pressure"
	pressureSourceWorkflow    = ".github/workflows/control-plane-performance.yml"
	pressureSummaryArtifact   = "control-plane-local-pressure"
	pressureCandidateArtifact = "control-plane-pressure-closed-loop-candidates"
	pressureLiveStatus        = "measured"
	pressureCandidateTTLHours = 24
)
