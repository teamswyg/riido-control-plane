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
		event := "provider_health_snapshot"
		if previousOK && found {
			previousStatus = old.HealthStatus
			event = "provider_health_changed"
		}
		log.Printf(
			"event=%s device_ref=%q provider=%q previous_status=%q status=%q diagnostic_code=%q failure_stage=%q diagnostic_summary=%q",
			event, providerHealthDeviceRef(next.DeviceID), runtime.Kind, previousStatus, runtime.HealthStatus,
			runtime.DiagnosticCode, providerDiagnosticStage(runtime.DiagnosticCode),
			canonicalProviderDiagnosticSummary(runtime.DiagnosticCode),
		)
	}
}

func traceProviderHealthChanges(ctx context.Context, previous DeviceRecord, previousOK bool, next DeviceRecord) {
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
		traceProviderHealthObservation(ctx, next.DeviceID, runtime, old.HealthStatus, previousOK && found)
	}
}

func traceProviderHealth(ctx context.Context, deviceID string, runtimes []RuntimeRecord) {
	for _, runtime := range runtimes {
		if runtime.HealthStatus != hostintegration.ProviderHealthHealthy {
			traceProviderHealthObservation(ctx, deviceID, runtime, "", false)
		}
	}
}

func traceProviderHealthObservation(ctx context.Context, deviceID string, runtime RuntimeRecord, previous hostintegration.ProviderHealthStatus, changed bool) {
	_, span := StartTraceSpan(ctx, nil, TraceSpanStart{
		Name: "provider health observation", Kind: TraceSpanKindInternal,
		Attributes: []TraceAttribute{
			StringTraceAttribute("riido.daemon.device_ref", providerHealthDeviceRef(deviceID)),
			StringTraceAttribute("riido.provider.kind", string(runtime.Kind)),
			StringTraceAttribute("riido.provider.health_status", string(runtime.HealthStatus)),
			StringTraceAttribute("riido.provider.previous_health_status", string(previous)),
			BoolTraceAttribute("riido.provider.health_changed", changed),
			StringTraceAttribute("riido.provider.diagnostic_code", string(runtime.DiagnosticCode)),
			StringTraceAttribute("riido.provider.diagnostic_summary", canonicalProviderDiagnosticSummary(runtime.DiagnosticCode)),
		},
	})
	if runtime.HealthStatus == hostintegration.ProviderHealthUnknown || runtime.HealthStatus == hostintegration.ProviderHealthUnavailable {
		span.RecordError(errors.New("provider health is not operational"))
	}
	span.End()
}
