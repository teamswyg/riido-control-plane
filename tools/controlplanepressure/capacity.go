package main

type capacityEstimate struct {
	Scenario           string  `json:"scenario"`
	MaxConcurrentUsers int     `json:"max_concurrent_users"`
	OpsPerSec          float64 `json:"ops_per_sec"`
	P95LatencyUS       int64   `json:"p95_latency_us"`
	AllocBytesPerOp    float64 `json:"alloc_bytes_per_op"`
	CPUSecondsPerOp    float64 `json:"cpu_seconds_per_op"`
	GoroutineDelta     int     `json:"goroutine_delta"`
	ErrorFree          bool    `json:"error_free"`
}

func capacityEstimates(runs []pressureRun) []capacityEstimate {
	best := map[string]pressureRun{}
	for _, run := range runs {
		if run.Errors != 0 {
			continue
		}
		current, ok := best[run.Scenario]
		if !ok || run.ConcurrentUsers > current.ConcurrentUsers {
			best[run.Scenario] = run
		}
	}
	out := make([]capacityEstimate, 0, len(best))
	for _, run := range runs {
		bestRun, ok := best[run.Scenario]
		if !ok || bestRun.ConcurrentUsers != run.ConcurrentUsers {
			continue
		}
		out = append(out, capacityEstimateFromRun(bestRun))
		delete(best, run.Scenario)
	}
	return out
}

func capacityEstimateFromRun(run pressureRun) capacityEstimate {
	return capacityEstimate{
		Scenario:           run.Scenario,
		MaxConcurrentUsers: run.ConcurrentUsers,
		OpsPerSec:          run.OpsPerSec,
		P95LatencyUS:       run.LatencyUS.P95,
		AllocBytesPerOp:    run.Resources.TotalAllocPerOp,
		CPUSecondsPerOp:    run.Resources.CPUSecondsPerOp,
		GoroutineDelta:     run.Resources.Goroutines,
		ErrorFree:          run.Errors == 0,
	}
}
