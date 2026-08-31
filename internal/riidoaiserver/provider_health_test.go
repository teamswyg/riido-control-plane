package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

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
