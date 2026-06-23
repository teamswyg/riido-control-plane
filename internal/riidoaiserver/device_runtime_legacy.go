package riidoaiserver

import "strings"

// legacyDaemonIDPrefix is the hardcoded daemon id used before per-machine UUIDs.
const legacyDaemonIDPrefix = "agentd-local:"

func pruneLegacyRuntimeRecords(devices []DeviceRecord) []DeviceRecord {
	for i := range devices {
		if len(devices[i].Runtimes) == 0 {
			continue
		}
		kept := devices[i].Runtimes[:0]
		for _, runtime := range devices[i].Runtimes {
			if strings.HasPrefix(strings.TrimSpace(runtime.RuntimeID), legacyDaemonIDPrefix) {
				continue
			}
			kept = append(kept, runtime)
		}
		devices[i].Runtimes = kept
	}
	return devices
}
