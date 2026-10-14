package riidoaiserver

import "testing"

func TestAppendRetainedClientReplayEventKeepsLatestEvents(t *testing.T) {
	events := make([]ClientStreamEvent, aiAgentClientReplayEventLimit)
	for i := range events {
		events[i] = ClientStreamEvent{Seq: int64(i + 1)}
	}
	got := appendRetainedClientReplayEvent(events, ClientStreamEvent{Seq: aiAgentClientReplayEventLimit + 1})
	if len(got) != aiAgentClientReplayEventLimit {
		t.Fatalf("event count = %d", len(got))
	}
	if got[0].Seq != 2 || got[len(got)-1].Seq != aiAgentClientReplayEventLimit+1 {
		t.Fatalf("retained seq range = %d..%d", got[0].Seq, got[len(got)-1].Seq)
	}
}

func TestAppendRetainedClientReplayEventAppendsBeforeLimit(t *testing.T) {
	got := appendRetainedClientReplayEvent(
		[]ClientStreamEvent{{Seq: 1}},
		ClientStreamEvent{Seq: 2},
	)
	if len(got) != 2 {
		t.Fatalf("event count = %d", len(got))
	}
	if got[0].Seq != 1 || got[1].Seq != 2 {
		t.Fatalf("events = %+v", got)
	}
}
