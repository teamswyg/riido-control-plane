package riidoaiserver

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"strings"

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

func traceProviderHealth(ctx context.Context, deviceID string, runtimes []RuntimeRecord) {
	for _, runtime := range runtimes {
		if runtime.HealthStatus == hostintegration.ProviderHealthHealthy {
			continue
		}
		_, span := StartTraceSpan(ctx, nil, TraceSpanStart{
			Name: "provider health observation",
			Kind: TraceSpanKindInternal,
			Attributes: []TraceAttribute{
				StringTraceAttribute("riido.daemon.device_ref", providerHealthDeviceRef(deviceID)),
				StringTraceAttribute("riido.provider.kind", string(runtime.Kind)),
				StringTraceAttribute("riido.provider.health_status", string(runtime.HealthStatus)),
				StringTraceAttribute("riido.provider.diagnostic_code", string(runtime.DiagnosticCode)),
				StringTraceAttribute("riido.provider.diagnostic_summary", runtime.DiagnosticSummary),
			},
		})
		if runtime.HealthStatus == hostintegration.ProviderHealthUnknown ||
			runtime.HealthStatus == hostintegration.ProviderHealthUnavailable {
			span.RecordError(errors.New("provider health is not operational"))
		}
		span.End()
	}
}

func providerHealthDeviceRef(deviceID string) string {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return "unknown"
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(deviceID)))[:12]
}

func providerDiagnosticStage(code hostintegration.ProviderDiagnosticCode) string {
	switch code {
	case hostintegration.ProviderDiagnosticExecutableMissing:
		return "executable"
	case hostintegration.ProviderDiagnosticLoginRequired, hostintegration.ProviderDiagnosticAuthProbeFailed:
		return "authentication"
	case hostintegration.ProviderDiagnosticVersionUnsupported, hostintegration.ProviderDiagnosticVersionProbeFailed:
		return "version"
	case hostintegration.ProviderDiagnosticCapabilityProbeFailed:
		return "capability"
	case hostintegration.ProviderDiagnosticRuntimeError:
		return "runtime"
	case hostintegration.ProviderDiagnosticProbeFailed:
		return "probe"
	default:
		return "none"
	}
}
