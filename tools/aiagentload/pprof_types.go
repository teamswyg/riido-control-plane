package main

type pprofEvidence struct {
	Enabled        bool          `json:"enabled"`
	BaseHost       string        `json:"base_host,omitempty"`
	ProfileSeconds int           `json:"profile_seconds,omitempty"`
	Samples        []pprofSample `json:"samples,omitempty"`
}

type pprofSample struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	Status        int    `json:"status,omitempty"`
	Bytes         int64  `json:"bytes"`
	LatencyMs     int64  `json:"latency_ms"`
	ErrorCategory string `json:"error_category,omitempty"`
}

type pprofTarget struct {
	name string
	path string
}
