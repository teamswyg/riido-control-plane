package riidoaiserver

import "strings"

func backfillRuntimeProviderVersionsFromSeed(devices, seed []DeviceRecord) []DeviceRecord {
	versions := seedRuntimeProviderVersions(seed)
	if len(versions) == 0 {
		return devices
	}
	for deviceIndex := range devices {
		for runtimeIndex := range devices[deviceIndex].Runtimes {
			runtime := &devices[deviceIndex].Runtimes[runtimeIndex]
			if strings.TrimSpace(runtime.ProviderVersion) != "" {
				continue
			}
			if version := versions[strings.TrimSpace(runtime.RuntimeID)]; version != "" {
				runtime.ProviderVersion = version
			}
		}
	}
	return devices
}

func seedRuntimeProviderVersions(devices []DeviceRecord) map[string]string {
	versions := map[string]string{}
	for _, device := range devices {
		for _, runtime := range device.Runtimes {
			runtimeID := strings.TrimSpace(runtime.RuntimeID)
			version := strings.TrimSpace(runtime.ProviderVersion)
			if runtimeID == "" || version == "" {
				continue
			}
			versions[runtimeID] = version
		}
	}
	return versions
}
