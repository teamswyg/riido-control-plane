package riidoaiserver

func applyHTTPClientSurfaceTotals(snapshot *MetricsSnapshot, totals map[string]int64) {
	if snapshot == nil {
		return
	}
	snapshot.HTTPRequestsDaemonTotal = totals["daemon"]
	snapshot.HTTPRequestsClientAppTotal = totals["client_app"]
	snapshot.HTTPRequestsDesktopTotal = totals["desktop"]
	snapshot.HTTPRequestsDesktopCandidateTotal = totals["desktop_candidate"]
	snapshot.HTTPRequestsComponentIntegrationTotal = totals["component_integration"]
	snapshot.HTTPRequestsUnknownSurfaceTotal = totals["unknown"]
}
