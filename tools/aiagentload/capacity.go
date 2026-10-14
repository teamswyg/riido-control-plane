package main

type capacityEstimate struct {
	ConcurrentUsers int     `json:"concurrent_users"`
	RequestsPerSec  float64 `json:"requests_per_second"`
	SuccessPerSec   float64 `json:"success_per_second"`
	P95LatencyMs    int64   `json:"p95_latency_ms"`
	FailureRatePct  float64 `json:"failure_rate_pct"`
	ErrorFree       bool    `json:"error_free"`
}

func capacityFromReport(r report) capacityEstimate {
	return capacityEstimate{
		ConcurrentUsers: r.Concurrency,
		RequestsPerSec:  r.RequestsPerSec,
		SuccessPerSec:   r.SuccessPerSec,
		P95LatencyMs:    r.Latency.P95,
		FailureRatePct:  r.FailureRatePct,
		ErrorFree:       r.Failures == 0,
	}
}
