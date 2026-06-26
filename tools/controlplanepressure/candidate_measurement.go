package main

func pressureCandidateMeasurements(capacity []capacityEstimate) map[string]pressureCandidateMeasurement {
	out := make(map[string]pressureCandidateMeasurement, len(capacity))
	for _, row := range capacity {
		out[row.Scenario] = pressureCandidateMeasurement{
			MaxConcurrentUsers: row.MaxConcurrentUsers,
			OpsPerSec:          row.OpsPerSec,
			P95LatencyUS:       row.P95LatencyUS,
			AllocBytesPerOp:    row.AllocBytesPerOp,
			CPUSecondsPerOp:    row.CPUSecondsPerOp,
			GoroutineDelta:     row.GoroutineDelta,
			ErrorFree:          row.ErrorFree,
		}
	}
	return out
}
