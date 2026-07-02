package main

type currentTaskVisualEvidence struct {
	Redacted     bool                      `json:"redacted"`
	Screenshot   currentTaskScreenshot     `json:"screenshot"`
	Metrics      currentTaskVisibleMetrics `json:"visible_metrics"`
	Observations []currentTaskObservation  `json:"observations"`
	NextArtifact string                    `json:"next_artifact"`
}

type currentTaskScreenshot struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type currentTaskVisibleMetrics struct {
	Thinking   int `json:"thinking_count"`
	QueuedCopy int `json:"queued_copy_count"`
}

type currentTaskObservation struct {
	Status   string `json:"status"`
	Decision string `json:"decision"`
}
