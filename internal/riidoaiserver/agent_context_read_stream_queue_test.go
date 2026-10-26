package riidoaiserver

import (
	"context"
	"testing"
	"time"
)

func TestContextReadStreamQueueBoundsWorkAndCancelsWorker(t *testing.T) {
	started, stopped := make(chan struct{}), make(chan struct{})
	q := contextReadDeliveryQueue{ctx: t.Context(), results: make(chan contextReadDeliveryResult, 1), filters: []clientEventDeliveryFilter{
		func(ctx context.Context, event ClientStreamEvent) (ClientStreamEvent, bool) {
			close(started)
			<-ctx.Done()
			close(stopped)
			return event, false
		},
	}}
	defer q.stop()
	for seq := range 1000 {
		q.enqueue(ClientStreamEvent{Seq: int64(seq + 1), EventType: AgentClientEventContextRead})
	}
	<-started
	if !q.active || len(q.pending) != 32 {
		t.Fatalf("metadata validation queue grew: active=%v pending=%d", q.active, len(q.pending))
	}
	q.supersede(1001)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("context-aware metadata worker leaked after terminal delivery")
	}
	result := <-q.results
	if _, allowed := q.complete(result); allowed || q.active || len(q.pending) != 0 {
		t.Fatal("superseded metadata was retained or restarted")
	}
}

func TestContextReadStreamDropsResultReadyBeforeCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	q := contextReadDeliveryQueue{ctx: ctx}
	result := contextReadDeliveryResult{event: ClientStreamEvent{Seq: 1}, allowed: true}
	cancel() // Both result and request cancellation may be ready in the select.
	if _, allowed := q.complete(result); allowed {
		t.Fatal("ready metadata result must not write after subscriber cancellation")
	}
}
