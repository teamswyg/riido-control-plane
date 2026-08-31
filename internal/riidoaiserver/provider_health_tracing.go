package riidoaiserver

import (
	"context"
	"errors"
	"log"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func logProviderHealthChanges(previous DeviceRecord, previousOK bool, next DeviceRecord) {
	previousByID := make(map[string]RuntimeRecord, len(previous.Runtimes))
	for _, runtime := range previous.Runtimes {
		previousByID[runtime.RuntimeID] = runtime
	}
	for _, runtime := range next.Runtimes {
		old, found := previousByID[runtime.RuntimeID]
		if found && old.HealthStatus == runtime.HealthStatus && old.DiagnosticCode == runtime.DiagnosticCode {
			continue
		}
		if !found && runtime.HealthStatus == hostintegration.ProviderHealthHealthy {
			continue
		}
		previousStatus := hostintegration.ProviderHealthStatus("")
		if previousOK && found {
			previousStatus = old.HealthStatus
		}
		log.Printf(
			"event=provider_health_changed provider=%q previous_status=%q status=%q diagnostic_code=%q",
			runtime.Kind, previousStatus, runtime.HealthStatus, runtime.DiagnosticCode,
		)
	}
}

func traceProviderHealth(ctx context.Context, runtimes []RuntimeRecord) {
	for _, runtime := range runtimes {
		if runtime.HealthStatus == hostintegration.ProviderHealthHealthy {
			continue
		}
		_, span := StartTraceSpan(ctx, nil, TraceSpanStart{
			Name: "provider health observation",
			Kind: TraceSpanKindInternal,
			Attributes: []TraceAttribute{
				StringTraceAttribute("riido.provider.kind", string(runtime.Kind)),
				StringTraceAttribute("riido.provider.health_status", string(runtime.HealthStatus)),
				StringTraceAttribute("riido.provider.diagnostic_code", string(runtime.DiagnosticCode)),
				StringTraceAttribute("riido.provider.diagnostic_summary", runtime.DiagnosticSummary),
			},
		})
		if runtime.HealthStatus == hostintegration.ProviderHealthUnknown {
			span.RecordError(errors.New("provider health is unknown"))
		}
		span.End()
	}
}
