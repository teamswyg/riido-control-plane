package riidoaiserver

import "sort"

func sortStoreOperationMetrics(operations []StoreOperationMetric) {
	sort.Slice(operations, func(i, j int) bool {
		if operations[i].CallsTotal != operations[j].CallsTotal {
			return operations[i].CallsTotal > operations[j].CallsTotal
		}
		return operations[i].Operation < operations[j].Operation
	})
}
