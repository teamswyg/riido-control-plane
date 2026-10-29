package riidoaiserver

import (
	"github.com/teamswyg/riido-contracts/hostintegration"
	"strings"
	"testing"
)

func TestRuntimeSnapshotInvalidVersionDoesNotRejectHeartbeat(t *testing.T) {
	for _, version := range []string{"Warning: update\ncodex-cli 0.161.0", strings.Repeat("x", 129), "\x1b[31m2.1.3"} {
		runtimes, err := normalizeRuntimeSnapshotRecords("device", "owner", "daemon", "production", []RuntimeSnapshotRecord{{RuntimeID: "device:codex", Kind: RuntimeKindCodex, ProviderVersion: version, Availability: RuntimeAvailabilityOnline, DetectionState: RuntimeDetectionStateDetected}})
		if err != nil {
			t.Fatal(err)
		}
		got := runtimes[0]
		if got.ProviderVersion != "" || got.Availability != RuntimeAvailabilityOnline || got.HealthStatus != hostintegration.ProviderHealthUnknown || got.DiagnosticCode != hostintegration.ProviderDiagnosticVersionProbeFailed {
			t.Fatalf("unexpected normalized runtime: %+v", got)
		}
	}
}
