package riidoaiserver

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestLogProviderHealthChangesOnlyLogsTransitions(t *testing.T) {
	var output bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previousWriter) })

	unknown := RuntimeRecord{
		RuntimeID:         "runtime-1",
		Kind:              RuntimeKindCodex,
		HealthStatus:      hostintegration.ProviderHealthUnknown,
		DiagnosticCode:    hostintegration.ProviderDiagnosticProbeFailed,
		DiagnosticSummary: "sensitive detail",
	}
	logProviderHealthChanges(DeviceRecord{}, false, DeviceRecord{Runtimes: []RuntimeRecord{unknown}})
	logProviderHealthChanges(DeviceRecord{Runtimes: []RuntimeRecord{unknown}}, true, DeviceRecord{Runtimes: []RuntimeRecord{unknown}})
	recovered := unknown
	recovered.HealthStatus = hostintegration.ProviderHealthHealthy
	recovered.DiagnosticCode = hostintegration.ProviderDiagnosticNone
	logProviderHealthChanges(DeviceRecord{Runtimes: []RuntimeRecord{unknown}}, true, DeviceRecord{Runtimes: []RuntimeRecord{recovered}})

	got := output.String()
	if strings.Count(got, "event=provider_health_changed") != 2 {
		t.Fatalf("provider health transition logs = %q", got)
	}
	if !strings.Contains(got, `status="unknown" diagnostic_code="probe-failed"`) ||
		!strings.Contains(got, `previous_status="unknown" status="healthy" diagnostic_code="none"`) {
		t.Fatalf("provider health transition logs = %q", got)
	}
	if strings.Contains(got, unknown.DiagnosticSummary) {
		t.Fatalf("provider health log leaked diagnostic summary: %q", got)
	}
}

func TestNormalizeRuntimeProviderHealthKeepsUnknownDiagnosticBounded(t *testing.T) {
	health, code, summary, err := normalizeRuntimeProviderHealth(RuntimeSnapshotRecord{
		HealthStatus:      hostintegration.ProviderHealthUnknown,
		DiagnosticCode:    hostintegration.ProviderDiagnosticProbeFailed,
		DiagnosticSummary: "provider probe did not complete",
	})
	if err != nil {
		t.Fatal(err)
	}
	if health != hostintegration.ProviderHealthUnknown || code != hostintegration.ProviderDiagnosticProbeFailed || summary != "provider probe did not complete" {
		t.Fatalf("health=%q code=%q summary=%q", health, code, summary)
	}
}

func TestNormalizeRuntimeProviderHealthFillsUnknownDiagnostic(t *testing.T) {
	health, code, summary, err := normalizeRuntimeProviderHealth(RuntimeSnapshotRecord{
		HealthStatus: hostintegration.ProviderHealthUnknown,
	})
	if err != nil {
		t.Fatal(err)
	}
	if health != hostintegration.ProviderHealthUnknown || code != hostintegration.ProviderDiagnosticProbeFailed || summary != "provider probe did not complete" {
		t.Fatalf("health=%q code=%q summary=%q", health, code, summary)
	}
}

func TestNormalizeRuntimeProviderHealthRejectsRawDiagnostic(t *testing.T) {
	_, _, _, err := normalizeRuntimeProviderHealth(RuntimeSnapshotRecord{
		HealthStatus:      hostintegration.ProviderHealthUnknown,
		DiagnosticCode:    hostintegration.ProviderDiagnosticProbeFailed,
		DiagnosticSummary: "bearer secret and /Users/private/path",
	})
	if err == nil {
		t.Fatal("raw diagnostic summary must be rejected")
	}
}

func TestTraceProviderHealthRecordsUnknownDiagnostic(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	traceProviderHealth(ctx, []RuntimeRecord{{
		Kind:              RuntimeKindCodex,
		HealthStatus:      hostintegration.ProviderHealthUnknown,
		DiagnosticCode:    hostintegration.ProviderDiagnosticProbeFailed,
		DiagnosticSummary: "provider probe did not complete",
	}})
	spans := recorder.snapshot()
	if len(spans) != 1 || len(spans[0].Errors) != 1 {
		t.Fatalf("provider health spans = %+v", spans)
	}
	if got := spans[0].Attributes["riido.provider.diagnostic_code"]; got != "probe-failed" {
		t.Fatalf("diagnostic code = %q", got)
	}
}
