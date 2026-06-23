package riidoaiserver

func storeOperationMetricsByName(metrics []StoreOperationMetric) map[string]StoreOperationMetric {
	out := make(map[string]StoreOperationMetric, len(metrics))
	for _, metric := range metrics {
		out[metric.Operation] = metric
	}
	return out
}
