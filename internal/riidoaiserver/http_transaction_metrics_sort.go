package riidoaiserver

import "sort"

func sortHTTPTransactionMetrics(transactions []HTTPTransactionMetric) {
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].RequestsTotal != transactions[j].RequestsTotal {
			return transactions[i].RequestsTotal > transactions[j].RequestsTotal
		}
		if transactions[i].Route != transactions[j].Route {
			return transactions[i].Route < transactions[j].Route
		}
		if transactions[i].ClientSurface != transactions[j].ClientSurface {
			return transactions[i].ClientSurface < transactions[j].ClientSurface
		}
		if transactions[i].Method != transactions[j].Method {
			return transactions[i].Method < transactions[j].Method
		}
		return transactions[i].StatusCode < transactions[j].StatusCode
	})
}
