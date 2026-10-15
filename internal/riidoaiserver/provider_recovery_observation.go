package riidoaiserver

import (
	"context"
	"errors"
	"log"

	"github.com/teamswyg/riido-contracts/assignment"
)

func observeProviderRecoveryEvent(ctx context.Context, event TaskEvent) {
	recovery := assignment.RecoveryContextFromMetadata(event.Metadata)
	if recovery == nil {
		return
	}
	traceID, spanID := traceCorrelationFromContext(ctx)
	log.Printf("event=provider_recovery_transition assignment_ref=%q event_ref=%q provider=%q recovery_state=%q recovery_reason=%q circuit=%q outcome_certainty=%q attempt=%d switch_count=%d trace_id=%q span_id=%q", providerHealthDeviceRef(event.AssignmentID), providerHealthDeviceRef(event.EventID), recovery.Provider, recovery.State, recovery.Reason, recovery.Circuit, recovery.OutcomeCertainty, recovery.Attempt, recovery.SwitchCount, traceID, spanID)
	_, span := StartTraceSpan(ctx, nil, TraceSpanStart{Name: "provider recovery transition", Kind: TraceSpanKindInternal, Attributes: []TraceAttribute{
		StringTraceAttribute("riido.provider.kind", recovery.Provider), StringTraceAttribute("riido.recovery.state", recovery.State), StringTraceAttribute("riido.recovery.reason", recovery.Reason), StringTraceAttribute("riido.recovery.circuit", recovery.Circuit), StringTraceAttribute("riido.recovery.outcome_certainty", recovery.OutcomeCertainty), Int64TraceAttribute("riido.recovery.attempt", int64(recovery.Attempt)), Int64TraceAttribute("riido.recovery.switch_count", int64(recovery.SwitchCount)),
	}})
	if recovery.State == "unavailable" || recovery.State == "unknown" {
		span.RecordError(errors.New("provider recovery did not verify availability"))
	}
	span.End()
}
