package main

type pressureEvidence struct {
	Runs     []pressureRun     `json:"runs"`
	Capacity []capacitySummary `json:"capacity_estimates"`
	Findings []pressureFinding `json:"findings"`
}

type pressureRun struct {
	Scenario  string            `json:"scenario"`
	Errors    int64             `json:"errors"`
	Resources pressureResources `json:"resource_delta"`
}

type pressureResources struct {
	Goroutines int `json:"goroutines"`
}

type capacitySummary struct {
	Scenario        string  `json:"scenario"`
	AllocBytesPerOp float64 `json:"alloc_bytes_per_op"`
	GoroutineDelta  int     `json:"goroutine_delta"`
	ErrorFree       bool    `json:"error_free"`
}

type pressureFinding struct {
	ID       string  `json:"id"`
	Scenario string  `json:"scenario"`
	Value    float64 `json:"value"`
}
