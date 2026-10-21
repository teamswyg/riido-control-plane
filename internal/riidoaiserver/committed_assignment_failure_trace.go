package riidoaiserver

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func traceCommittedAssignmentFailure(ctx context.Context, state *storeState, event TaskEvent) {
	if event.Type != EventAssignmentFailed || event.State == AssignmentCancelled || event.State == AssignmentCancelling {
		return
	}
	status := boundedResultStatus(event.Metadata[metadatakeys.AssignmentResultStatus.String()])
	if status == "cancelled" || status == "needs_input" {
		return
	}
	// This identifies the original assignment/lease scope. A switched execution's
	// actual provider is recorded separately by its recovery journal.
	provider := boundedProvider(state.assignments[event.AssignmentID].RuntimeProvider)
	_, span := StartTraceSpan(ctx, nil, TraceSpanStart{Name: "assignment failure committed", Kind: TraceSpanKindInternal, Attributes: []TraceAttribute{
		StringTraceAttribute("riido.assignment.provider", provider),
		StringTraceAttribute("riido.assignment.result_status", status),
		StringTraceAttribute("riido.assignment.failure_category", boundedFailureCategory(event.Metadata[metadatakeys.AssignmentFailureCategory.String()])),
		StringTraceAttribute("riido.assignment.ref", providerHealthDeviceRef(event.AssignmentID)),
	}})
	span.RecordError(errors.New("assignment execution failed"))
	span.End()
}
