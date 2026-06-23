package riidoaiserver

import "testing"

func TestPruneLegacyRuntimeRecordsDropsAgentdLocal(t *testing.T) {
	devices := []DeviceRecord{{
		DeviceID: "dev_x",
		Runtimes: []RuntimeRecord{
			{RuntimeID: "agentd-local:claude"},
			{RuntimeID: "e98eefcd-66b4-49e2-a1bf-1cab74749e2d:claude"},
			{RuntimeID: "agentd-local:codex"},
		},
	}}
	out := pruneLegacyRuntimeRecords(devices)
	if len(out[0].Runtimes) != 1 ||
		out[0].Runtimes[0].RuntimeID != "e98eefcd-66b4-49e2-a1bf-1cab74749e2d:claude" {
		t.Fatalf("legacy runtimes not pruned: %+v", out[0].Runtimes)
	}
}
