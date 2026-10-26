package riidoaiserver

import "context"

type contextReadDeliveryResult struct {
	event   ClientStreamEvent
	allowed bool
}

// One worker and at most 32 queued hints per subscriber. The writer never
// waits on this worker; newer progress supersedes old hints to keep cursor order.
type contextReadDeliveryQueue struct {
	ctx      context.Context
	filters  []clientEventDeliveryFilter
	results  chan contextReadDeliveryResult
	pending  []ClientStreamEvent
	cancel   context.CancelFunc
	active   bool
	lastSent int64
}

func (q *contextReadDeliveryQueue) enqueue(event ClientStreamEvent) {
	if len(q.pending) < 32 {
		q.pending = append(q.pending, event)
	}
	q.start()
}

func (q *contextReadDeliveryQueue) start() {
	if q.ctx.Err() != nil || q.active || len(q.pending) == 0 {
		return
	}
	event := q.pending[0]
	q.pending = q.pending[1:]
	ctx, cancel := context.WithCancel(q.ctx)
	q.cancel, q.active = cancel, true
	go func() {
		filtered, allowed := filterAIAgentClientDelivery(ctx, event, q.filters)
		q.results <- contextReadDeliveryResult{filtered, allowed && ctx.Err() == nil}
	}()
}

func (q *contextReadDeliveryQueue) supersede(seq int64) {
	q.lastSent = max(q.lastSent, seq)
	q.pending = nil
	q.stop()
}

func (q *contextReadDeliveryQueue) stop() {
	if q.cancel != nil {
		q.cancel()
	}
}

func (q *contextReadDeliveryQueue) complete(result contextReadDeliveryResult) (ClientStreamEvent, bool) {
	q.stop()
	q.active = false
	allowed := q.ctx.Err() == nil && result.allowed && result.event.Seq > q.lastSent
	if allowed {
		q.lastSent = result.event.Seq
	}
	q.start()
	return result.event, allowed
}
