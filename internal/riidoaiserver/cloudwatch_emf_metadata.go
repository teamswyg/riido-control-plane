package riidoaiserver

type cloudWatchEMFMetadata struct {
	Timestamp         int64                   `json:"Timestamp"`
	CloudWatchMetrics []cloudWatchMetricGroup `json:"CloudWatchMetrics"`
}

type cloudWatchMetricGroup struct {
	Namespace  string                 `json:"Namespace"`
	Dimensions [][]string             `json:"Dimensions"`
	Metrics    []cloudWatchMetricSpec `json:"Metrics"`
}

type cloudWatchMetricSpec struct {
	Name string `json:"Name"`
	Unit string `json:"Unit"`
}
