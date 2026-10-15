package riidoaiserver

import "context"

// Call only after persistence has accepted an event, never while constructing
// a speculative trial or retrying a failed conditional write.
func observeCommittedAssignmentEvent(ctx context.Context, state *storeState, event TaskEvent) {
	logAssignmentFailure(state, event, ctx)
	traceCommittedAssignmentFailure(ctx, state, event)
	observeProviderRecoveryEvent(ctx, event)
}
