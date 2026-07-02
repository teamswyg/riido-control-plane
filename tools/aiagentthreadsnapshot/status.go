package main

func isTerminal(state string) bool {
	switch state {
	case "completed", "failed", "stopped", "cancelled", "timeout":
		return true
	default:
		return false
	}
}

func decide(rep report) decisionSummary {
	if hasEndpointError(rep.Endpoints) {
		return decisionSummary{Status: "partial", Reason: "one or more API endpoints failed"}
	}
	if rep.V3.RunningCount > 0 || rep.V2.RunningCount > 0 || len(rep.SSEEvents) > 0 {
		return decisionSummary{Status: "captured", Reason: "snapshot includes active or live progress signal"}
	}
	return decisionSummary{Status: "baseline", Reason: "no active progress signal observed during capture window"}
}

func hasEndpointError(endpoints []endpointObservation) bool {
	for _, endpoint := range endpoints {
		if endpoint.Error != "" || endpoint.StatusCode == 0 {
			return true
		}
	}
	return false
}
