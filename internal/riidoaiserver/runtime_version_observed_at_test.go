package riidoaiserver

import (
	"testing"
	"time"
)

func TestVersionObservationMetadataCannotRejectHeartbeat(t *testing.T) {
	observed := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	for _, value := range []string{observed.Format(time.RFC3339), "invalid", time.Now().Add(time.Hour).Format(time.RFC3339), ""} {
		runtimes, err := normalizeRuntimeSnapshotRecords("device", "owner", "daemon", "production", []RuntimeSnapshotRecord{{RuntimeID: "device:codex", Kind: RuntimeKindCodex, ProviderVersion: "codex-cli 1.2.3", ProviderVersionObservedAt: value}})
		if err != nil {
			t.Fatal(err)
		}
		got := runtimes[0]
		if value == observed.Format(time.RFC3339) {
			if !got.ProviderVersionObservedAt.Equal(observed) {
				t.Fatal("observation lost")
			}
		} else if !got.ProviderVersionObservedAt.IsZero() {
			t.Fatal("invalid metadata accepted")
		}
		if got.LastDetectedAt.Before(observed.Add(time.Minute)) {
			t.Fatal("version time replaced liveness")
		}
	}
}
