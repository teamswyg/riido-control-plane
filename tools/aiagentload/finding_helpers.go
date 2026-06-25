package main

func endpointTotal(summary endpointSummary) float64 {
	return float64(summary.Total)
}

func endpointP95(summary endpointSummary) float64 {
	return float64(summary.Latency.P95)
}

func failureFinding(endpoints map[string]endpointSummary) (findingEntry, bool) {
	finding := endpointMaxFinding("failure_hot_endpoint", endpoints, "failures", "count", endpointFailures)
	return finding, finding.Value > 0
}

func endpointFailures(summary endpointSummary) float64 {
	return float64(summary.Failures)
}

func nextForMetric(metric string) string {
	switch metric {
	case "requests":
		return "Check this route first when request volume rises; keep route pattern metrics bounded."
	case "p95_latency_ms":
		return "Profile this route's handler/store calls before changing endpoint contracts."
	case "failures":
		return "Inspect status/error categories and promote recurring failures to a closed-loop verifier."
	default:
		return "Keep the finding as open-loop evidence until repeated observations justify a verifier."
	}
}
