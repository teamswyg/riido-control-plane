package riidoaiserver

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestProviderHealthTraceSuppressesUnchangedPollAndRecordsRecovery(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	ctx := WithTraceRecorder(context.Background(), recorder)
	store := NewDevelopmentAIAgentClientStore()
	principal := AuthorizationResult{PrincipalID: "user-health"}
	req := DeviceRuntimeSnapshotSyncRequest{DeviceID: "device-health", DaemonID: "daemon-health", Runtimes: []RuntimeSnapshotRecord{{RuntimeID: "runtime-health", Kind: RuntimeKindCodex, Availability: RuntimeAvailabilityOnline, DetectionState: "detected", HealthStatus: hostintegration.ProviderHealthUnknown, DiagnosticCode: hostintegration.ProviderDiagnosticProbeFailed}}}
	for range 2 {
		if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
			t.Fatal(err)
		}
	}
	req.Runtimes[0].HealthStatus = hostintegration.ProviderHealthHealthy
	req.Runtimes[0].DiagnosticCode = hostintegration.ProviderDiagnosticNone
	if _, err := store.SyncAIAgentDaemonRuntimeSnapshot(ctx, principal, req); err != nil {
		t.Fatal(err)
	}
	spans := recorder.snapshot()
	if len(spans) != 2 {
		t.Fatalf("health spans=%d want=2 (initial failure and recovery)", len(spans))
	}
	if spans[1].Attributes["riido.provider.health_status"] != "healthy" {
		t.Fatalf("recovery span missing: %+v", spans[1])
	}
}
