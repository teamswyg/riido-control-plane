package main

import (
	"net/http"
	"time"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func sampleMetricsSnapshot() riidoaiserver.MetricsSnapshot {
	return riidoaiserver.MetricsSnapshot{
		SchemaVersion:                    riidoaiserver.MetricsSchemaVersion,
		GeneratedAt:                      time.Unix(123, 456000000).UTC(),
		TasksTotal:                       2,
		AssignmentsTotal:                 3,
		AssignmentsByState:               assignmentStates(),
		PollRequestsTotal:                5,
		PollActionsTotal:                 pollActions(),
		AgentEventsTotal:                 7,
		TaskEventsTotal:                  11,
		SSESubscribers:                   13,
		SSEStreamsActive:                 2,
		SSEStreamsOpenedTotal:            5,
		SSEStreamsClosedTotal:            3,
		SSEStreamTTFBSamplesTotal:        5,
		SSEStreamTTFBMaxMilliseconds:     17,
		SSEStreamDurationSamplesTotal:    3,
		SSEStreamDurationMaxMilliseconds: 60000,
		SSEStreams: []riidoaiserver.SSEStreamMetric{{
			Route:              "/v2/client/workspaces/{workspace_id}/ai-agent/events",
			ClientSurface:      "client_app",
			StreamsOpenedTotal: 5,
			StreamsClosedTotal: 3,
			ActiveStreams:      2,
		}},
		OutboxErrorsTotal:                               17,
		EventAppendLatencySamplesTotal:                  19,
		EventAppendLatencyMaxMilliseconds:               89,
		HTTPRequestsTotal:                               23,
		HTTPResponsesByStatus:                           httpStatuses(),
		HTTPRequestLatencySamplesTotal:                  23,
		HTTPRequestLatencyMaxMilliseconds:               67,
		HTTPTransactions:                                httpTransactions(),
		StoreOperationCallsTotal:                        29,
		StoreOperationLatencySamplesTotal:               29,
		StoreOperationLatencyMaxMilliseconds:            78,
		StoreOperations:                                 storeOperations(),
		AIAgentClientSnapshotLoadCallsTotal:             31,
		AIAgentClientSnapshotLoadBytesLast:              40960,
		AIAgentClientSnapshotLoadLatencySamplesTotal:    31,
		AIAgentClientSnapshotLoadLatencyMaxMilliseconds: 81,
		AIAgentClientSnapshotSaveCallsTotal:             37,
		AIAgentClientSnapshotSaveBytesLast:              20480,
		AIAgentClientSnapshotSaveLatencySamplesTotal:    37,
		AIAgentClientSnapshotSaveLatencyMaxMilliseconds: 82,
	}
}

func httpStatuses() map[int]int64 {
	return map[int]int64{http.StatusOK: 17, http.StatusNotFound: 4, http.StatusInternalServerError: 2}
}
