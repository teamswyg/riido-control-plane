package main

import (
	"net/http"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func assignmentStates() map[riidoaiserver.AssignmentState]int {
	return map[riidoaiserver.AssignmentState]int{
		riidoaiserver.AssignmentQueued:  1,
		riidoaiserver.AssignmentRunning: 2,
	}
}

func pollActions() map[riidoaiserver.PollAction]int64 {
	return map[riidoaiserver.PollAction]int64{
		riidoaiserver.PollStart: 2,
		riidoaiserver.PollNone:  3,
	}
}

func httpTransactions() []riidoaiserver.HTTPTransactionMetric {
	return []riidoaiserver.HTTPTransactionMetric{{
		Method:                   http.MethodGet,
		Route:                    "/healthz",
		StatusCode:               http.StatusOK,
		RequestsTotal:            17,
		LatencySamplesTotal:      17,
		LatencyTotalMilliseconds: 221,
		LatencyMaxMilliseconds:   31,
		LatencyLastMilliseconds:  4,
	}}
}

func storeOperations() []riidoaiserver.StoreOperationMetric {
	return []riidoaiserver.StoreOperationMetric{{
		Operation:                riidoaiserver.StoreOperationPollAssignment.String(),
		CallsTotal:               21,
		LatencySamplesTotal:      21,
		LatencyTotalMilliseconds: 321,
		LatencyMaxMilliseconds:   78,
	}}
}
