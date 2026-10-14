package riidoaiserver

func retainLatestClientReplayEvents(events []ClientStreamEvent) []ClientStreamEvent {
	if len(events) <= aiAgentClientReplayEventLimit {
		return events
	}
	retained := make([]ClientStreamEvent, aiAgentClientReplayEventLimit)
	copy(retained, events[len(events)-aiAgentClientReplayEventLimit:])
	return retained
}

// retainLatestThreadProgressLines bounds the progress lines persisted per task
// thread. Lines accumulate per run and are serialized into the single AI Agent
// client snapshot DynamoDB item (400 KB hard limit); left uncapped they
// eventually make every client-projection write fail. Only the latest lines are
// kept for replay. Live progress is best-effort under backpressure, while terminal
// progress is prioritized and repaired from the durable thread projection.
