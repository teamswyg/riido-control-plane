package riidoaiserver

func applyCloudWatchEMFRuntimeEvents(envelope *cloudWatchEMFEnvelope, snapshot MetricsSnapshot) {
	envelope.PollRequestsTotal = snapshot.PollRequestsTotal
	envelope.PollNoneTotal = snapshot.PollActionsTotal[PollNone]
	envelope.PollStartTotal = snapshot.PollActionsTotal[PollStart]
	envelope.PollCancelTotal = snapshot.PollActionsTotal[PollCancel]
	envelope.PollActiveTotal = snapshot.PollActionsTotal[PollActive]
	envelope.AgentEventsTotal = snapshot.AgentEventsTotal
	envelope.TaskEventsTotal = snapshot.TaskEventsTotal
	envelope.SSESubscribers = snapshot.SSESubscribers
	envelope.SSEStreamsActive = snapshot.SSEStreamsActive
	envelope.SSEStreamsOpenedTotal = snapshot.SSEStreamsOpenedTotal
	envelope.SSEStreamsClosedTotal = snapshot.SSEStreamsClosedTotal
	envelope.SSEStreamTTFBSamplesTotal = snapshot.SSEStreamTTFBSamplesTotal
	envelope.SSEStreamTTFBTotalMilliseconds = snapshot.SSEStreamTTFBTotalMilliseconds
	envelope.SSEStreamTTFBMaxMilliseconds = snapshot.SSEStreamTTFBMaxMilliseconds
	envelope.SSEStreamTTFBLastMilliseconds = snapshot.SSEStreamTTFBLastMilliseconds
	envelope.SSEStreamDurationSamplesTotal = snapshot.SSEStreamDurationSamplesTotal
	envelope.SSEStreamDurationTotalMilliseconds = snapshot.SSEStreamDurationTotalMilliseconds
	envelope.SSEStreamDurationMaxMilliseconds = snapshot.SSEStreamDurationMaxMilliseconds
	envelope.SSEStreamDurationLastMilliseconds = snapshot.SSEStreamDurationLastMilliseconds
	envelope.SSEStreams = snapshot.SSEStreams
	envelope.OutboxErrorsTotal = snapshot.OutboxErrorsTotal
	envelope.EventAppendLatencySamplesTotal = snapshot.EventAppendLatencySamplesTotal
	envelope.EventAppendLatencyTotalMilliseconds = snapshot.EventAppendLatencyTotalMilliseconds
	envelope.EventAppendLatencyMaxMilliseconds = snapshot.EventAppendLatencyMaxMilliseconds
	envelope.EventAppendLatencyLastMilliseconds = snapshot.EventAppendLatencyLastMilliseconds
}
