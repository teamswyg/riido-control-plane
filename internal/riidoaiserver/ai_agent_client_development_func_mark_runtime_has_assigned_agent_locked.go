package riidoaiserver

func markRuntimeHasAssignedAgentLocked(devices []DeviceRecord, runtimeID string, value bool) {
	for deviceIndex := range devices {
		for runtimeIndex := range devices[deviceIndex].Runtimes {
			if devices[deviceIndex].Runtimes[runtimeIndex].RuntimeID == runtimeID {
				devices[deviceIndex].Runtimes[runtimeIndex].HasAssignedAgent = value
				return
			}
		}
	}
}
