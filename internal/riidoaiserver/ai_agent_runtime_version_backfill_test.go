package riidoaiserver

import "testing"

func TestBackfillRuntimeProviderVersionsDoesNotOverwriteSnapshotValue(t *testing.T) {
	devices := []DeviceRecord{{
		Runtimes: []RuntimeRecord{{
			RuntimeID:       "runtime-codex-dev",
			ProviderVersion: "codex-cli live",
		}},
	}}
	seed := []DeviceRecord{{
		Runtimes: []RuntimeRecord{{
			RuntimeID:       "runtime-codex-dev",
			ProviderVersion: "codex-cli seed",
		}},
	}}

	got := backfillRuntimeProviderVersionsFromSeed(devices, seed)
	if got[0].Runtimes[0].ProviderVersion != "codex-cli live" {
		t.Fatalf("provider_version was overwritten: %+v", got[0].Runtimes[0])
	}
}

func TestBackfillRuntimeProviderVersionsIgnoresUnknownRuntime(t *testing.T) {
	devices := []DeviceRecord{{Runtimes: []RuntimeRecord{{RuntimeID: "runtime-new"}}}}
	seed := []DeviceRecord{{Runtimes: []RuntimeRecord{{RuntimeID: "runtime-seed", ProviderVersion: "seed"}}}}

	got := backfillRuntimeProviderVersionsFromSeed(devices, seed)
	if got[0].Runtimes[0].ProviderVersion != "" {
		t.Fatalf("unknown runtime received guessed provider_version: %+v", got[0].Runtimes[0])
	}
}
