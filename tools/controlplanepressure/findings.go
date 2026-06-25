package main

type findingEntry struct {
	ID          string         `json:"id"`
	Scenario    string         `json:"scenario"`
	Concurrency int            `json:"concurrency"`
	Metric      string         `json:"metric"`
	Value       float64        `json:"value"`
	Unit        string         `json:"unit"`
	Next        string         `json:"next"`
	Candidate   candidateEntry `json:"candidate"`
}

func pressureFindings(runs []pressureRun) []findingEntry {
	return []findingEntry{
		maxFinding("allocation_hotspot", runs, "total_alloc_bytes_per_op", "bytes/op", allocPerOp),
		maxFinding("latency_hotspot", runs, "p95_latency_us", "microseconds", p95LatencyUS),
		maxFinding("cpu_hotspot", runs, "cpu_seconds_per_op", "seconds/op", cpuPerOp),
		maxFinding("goroutine_delta_max", runs, "goroutine_delta", "goroutines", goroutineDelta),
	}
}

func maxFinding(id string, runs []pressureRun, metric, unit string, value func(pressureRun) float64) findingEntry {
	var best pressureRun
	var bestValue float64
	for i, run := range runs {
		current := value(run)
		if i == 0 || current > bestValue {
			best, bestValue = run, current
		}
	}
	return findingEntry{
		ID: id, Scenario: best.Scenario, Concurrency: best.Concurrency, Metric: metric,
		Value: bestValue, Unit: unit, Next: best.Candidate.Next,
		Candidate: best.Candidate,
	}
}

func allocPerOp(run pressureRun) float64 {
	return run.Resources.TotalAllocPerOp
}

func p95LatencyUS(run pressureRun) float64 {
	return float64(run.LatencyUS.P95)
}

func cpuPerOp(run pressureRun) float64 {
	return run.Resources.CPUSecondsPerOp
}

func goroutineDelta(run pressureRun) float64 {
	return float64(run.Resources.Goroutines)
}
