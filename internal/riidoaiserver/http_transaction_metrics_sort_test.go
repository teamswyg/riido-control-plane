package riidoaiserver

import "testing"

func TestSortHTTPTransactionMetricsOrder(t *testing.T) {
	metrics := []HTTPTransactionMetric{
		{Method: "POST", Route: "/b", ClientSurface: "front", StatusCode: 200, RequestsTotal: 5},
		{Method: "GET", Route: "/a", ClientSurface: "daemon", StatusCode: 500, RequestsTotal: 5},
		{Method: "GET", Route: "/a", ClientSurface: "daemon", StatusCode: 200, RequestsTotal: 5},
		{Method: "POST", Route: "/a", ClientSurface: "daemon", StatusCode: 200, RequestsTotal: 5},
		{Method: "GET", Route: "/a", ClientSurface: "front", StatusCode: 200, RequestsTotal: 5},
		{Method: "GET", Route: "/z", ClientSurface: "front", StatusCode: 200, RequestsTotal: 9},
	}
	sortHTTPTransactionMetrics(metrics)
	want := []HTTPTransactionMetric{
		{Method: "GET", Route: "/z", ClientSurface: "front", StatusCode: 200, RequestsTotal: 9},
		{Method: "GET", Route: "/a", ClientSurface: "daemon", StatusCode: 200, RequestsTotal: 5},
		{Method: "GET", Route: "/a", ClientSurface: "daemon", StatusCode: 500, RequestsTotal: 5},
		{Method: "POST", Route: "/a", ClientSurface: "daemon", StatusCode: 200, RequestsTotal: 5},
		{Method: "GET", Route: "/a", ClientSurface: "front", StatusCode: 200, RequestsTotal: 5},
		{Method: "POST", Route: "/b", ClientSurface: "front", StatusCode: 200, RequestsTotal: 5},
	}
	for i := range want {
		if metrics[i] != want[i] {
			t.Fatalf("metric %d = %+v, want %+v; all=%+v", i, metrics[i], want[i], metrics)
		}
	}
}
