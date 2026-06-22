package riidoaiserver

func applyCloudWatchEMFHTTP(envelope *cloudWatchEMFEnvelope, snapshot MetricsSnapshot) {
	envelope.HTTPRequestsTotal = snapshot.HTTPRequestsTotal
	envelope.HTTPResponse2xxTotal = countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 2)
	envelope.HTTPResponse3xxTotal = countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 3)
	envelope.HTTPResponse4xxTotal = countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 4)
	envelope.HTTPResponse5xxTotal = countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 5)
	envelope.HTTPRequestLatencySamplesTotal = snapshot.HTTPRequestLatencySamplesTotal
	envelope.HTTPRequestLatencyTotalMilliseconds = snapshot.HTTPRequestLatencyTotalMilliseconds
	envelope.HTTPRequestLatencyMaxMilliseconds = snapshot.HTTPRequestLatencyMaxMilliseconds
	envelope.HTTPRequestLatencyLastMilliseconds = snapshot.HTTPRequestLatencyLastMilliseconds
	envelope.HTTPTransactions = snapshot.HTTPTransactions
}

func countHTTPStatusClass(statuses map[int]int64, class int) int64 {
	var total int64
	for statusCode, count := range statuses {
		if statusCode/100 == class {
			total += count
		}
	}
	return total
}
