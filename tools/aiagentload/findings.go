package main

type findingEntry struct {
	ID       string  `json:"id"`
	Endpoint string  `json:"endpoint,omitempty"`
	Metric   string  `json:"metric"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Next     string  `json:"next"`
}

func loadFindings(endpoints map[string]endpointSummary, resource resourceDelta) []findingEntry {
	rows := []findingEntry{
		endpointMaxFinding("volume_hot_endpoint", endpoints, "requests", "count", endpointTotal),
		endpointMaxFinding("latency_hot_endpoint", endpoints, "p95_latency_ms", "milliseconds", endpointP95),
		{
			ID:     "load_generator_allocation",
			Metric: "total_alloc_bytes_per_request",
			Value:  resource.TotalAllocPerRequest,
			Unit:   "bytes/request",
			Next:   "If this grows under the same scenario, profile the load harness before blaming server memory.",
		},
	}
	if finding, ok := failureFinding(endpoints); ok {
		rows = append(rows, finding)
	}
	return rows
}

func endpointMaxFinding(
	id string,
	endpoints map[string]endpointSummary,
	metric string,
	unit string,
	value func(endpointSummary) float64,
) findingEntry {
	var bestPath string
	var bestValue float64
	first := true
	for path, summary := range endpoints {
		current := value(summary)
		if first || current > bestValue {
			first = false
			bestPath, bestValue = path, current
		}
	}
	return findingEntry{ID: id, Endpoint: bestPath, Metric: metric, Value: bestValue, Unit: unit, Next: nextForMetric(metric)}
}
