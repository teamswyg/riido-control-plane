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
// kept for replay — live SSE subscribers still receive every line in real time.
