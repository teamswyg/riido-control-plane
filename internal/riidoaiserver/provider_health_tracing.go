package riidoaiserver

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

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
