package riidoaiserver

import (
	"net/http"
	"time"
)

func sampleCloudWatchEMFSnapshot() MetricsSnapshot {
	return MetricsSnapshot{
		SchemaVersion:                      MetricsSchemaVersion,
		GeneratedAt:                        time.Unix(123, 456000000).UTC(),
		TasksTotal:                         2,
		AssignmentsTotal:                   3,
		AssignmentsByState:                 map[AssignmentState]int{AssignmentQueued: 1, AssignmentRunning: 2},
		PollRequestsTotal:                  5,
		PollActionsTotal:                   map[PollAction]int64{PollStart: 2, PollNone: 3},
		AgentEventsTotal:                   7,
		TaskEventsTotal:                    11,
		SSESubscribers:                     13,
		SSEStreamsActive:                   2,
		SSEStreamsOpenedTotal:              5,
		SSEStreamsClosedTotal:              3,
		SSEStreamTTFBSamplesTotal:          5,
		SSEStreamTTFBTotalMilliseconds:     50,
		SSEStreamTTFBMaxMilliseconds:       17,
		SSEStreamTTFBLastMilliseconds:      9,
		SSEStreamDurationSamplesTotal:      3,
		SSEStreamDurationTotalMilliseconds: 123000,
		SSEStreamDurationMaxMilliseconds:   60000,
		SSEStreamDurationLastMilliseconds:  30000,
		SSEStreams: []SSEStreamMetric{{
			Route:              "/v2/client/workspaces/{workspace_id}/ai-agent/events",
			ClientSurface:      "client_app",
			StreamsOpenedTotal: 5,
			StreamsClosedTotal: 3,
			ActiveStreams:      2,
		}},
		OutboxErrorsTotal:                               17,
		EventAppendLatencySamplesTotal:                  19,
		EventAppendLatencyTotalMilliseconds:             230,
		EventAppendLatencyMaxMilliseconds:               89,
		EventAppendLatencyLastMilliseconds:              34,
		HTTPRequestsTotal:                               23,
		HTTPResponsesByStatus:                           map[int]int64{http.StatusOK: 17, http.StatusNotFound: 4, http.StatusInternalServerError: 2},
		HTTPRequestLatencySamplesTotal:                  23,
		HTTPRequestLatencyTotalMilliseconds:             345,
		HTTPRequestLatencyMaxMilliseconds:               67,
		HTTPRequestLatencyLastMilliseconds:              12,
		HTTPRequestsDaemonTotal:                         5,
		HTTPRequestsClientAppTotal:                      12,
		HTTPRequestsDesktopTotal:                        3,
		HTTPRequestsDesktopCandidateTotal:               1,
		HTTPRequestsComponentIntegrationTotal:           2,
		StoreOperationCallsTotal:                        29,
		StoreOperationErrorsTotal:                       3,
		StoreOperationLatencySamplesTotal:               29,
		StoreOperationLatencyTotalMilliseconds:          456,
		StoreOperationLatencyMaxMilliseconds:            78,
		StoreOperationLatencyLastMilliseconds:           9,
		AIAgentClientSnapshotLoadCallsTotal:             31,
		AIAgentClientSnapshotLoadBytesLast:              40960,
		AIAgentClientSnapshotLoadLatencySamplesTotal:    31,
		AIAgentClientSnapshotLoadLatencyMaxMilliseconds: 81,
		AIAgentClientSnapshotSaveCallsTotal:             37,
		AIAgentClientSnapshotSaveBytesLast:              20480,
		AIAgentClientSnapshotSaveLatencySamplesTotal:    37,
		AIAgentClientSnapshotSaveLatencyMaxMilliseconds: 82,
		HTTPTransactions:                                sampleCloudWatchEMFHTTPTransactions(),
		StoreOperations:                                 sampleCloudWatchEMFStoreOperations(),
	}
}

func sampleCloudWatchEMFHTTPTransactions() []HTTPTransactionMetric {
	return []HTTPTransactionMetric{{
		Method:                   http.MethodGet,
		Route:                    "/healthz",
		ClientSurface:            "daemon",
		StatusCode:               http.StatusOK,
		RequestsTotal:            17,
		LatencySamplesTotal:      17,
		LatencyTotalMilliseconds: 221,
		LatencyMaxMilliseconds:   31,
		LatencyLastMilliseconds:  4,
	}}
}

func sampleCloudWatchEMFStoreOperations() []StoreOperationMetric {
	return []StoreOperationMetric{{
		Operation:                StoreOperationPollAssignment.String(),
		CallsTotal:               21,
		ErrorsTotal:              2,
		LatencySamplesTotal:      21,
		LatencyTotalMilliseconds: 321,
		LatencyMaxMilliseconds:   78,
		LatencyLastMilliseconds:  9,
	}}
}
